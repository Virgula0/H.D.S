//nolint:all // test
package main

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/hcxtools"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/mandiant/gocat/v6"
	"github.com/mandiant/gocat/v6/hcargp"
)

const DebugTest = true

var gocatOptions = gocat.Options{
	SharedPath: "/usr/local/share/hashcat/OpenCL",
}

func goatCallback(resultsmap map[string]*string) gocat.EventCallback {
	return func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case gocat.LogPayload:
			if DebugTest {
				fmt.Printf("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.ActionPayload:
			if DebugTest {
				fmt.Printf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
			}
		case gocat.CrackedPayload:
			if DebugTest {
				fmt.Printf("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
			if resultsmap != nil {
				resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
			}
		case gocat.FinalStatusPayload:
			if DebugTest {
				fmt.Printf("FINAL STATUS -> %v\n", pl.Status)
			}
		case gocat.TaskInformationPayload:
			if DebugTest {
				fmt.Printf("TASK INFO -> %v\n", pl)
			}
		}
	}
}

func main() {

	_, err := environment.InitEnvironment()

	if err != nil {
		log.Fatal()
	}

	// start client
	client := grpcclient.InitClient()
	defer client.ClientCloser()

	//gui.InitLoginWindow(client)

	response, _ := client.Authenticate("admin", "test1234")
	*client.Credentials.JWT = response.Details

	go client.Authenticator() // run authenticator

	machineID, err := utils.MachineID()

	if err != nil {
		log.Panic(err.Error())
	}

	info, err := client.GetClientInfo(constants.MachineName, machineID)
	if err != nil {
		log.Panic(err.Error())
	}

	clientUUID := info.GetClientUuid()

	stream, err := client.HashcatChat()

	if err != nil {
		log.Panic(err.Error())
	}

	ticker := time.NewTicker(time.Second * 10)

	for {
		var handshake entities.Handshake
		for {
			log.Println("[CLIENT] Listening for tasks...")

			handshake = entities.Handshake{
				ClientUUID:       new(string),
				CrackedDate:      new(string),
				HashcatOptions:   new(string),
				HashcatLogs:      new(string),
				CrackedHandshake: new(string),
				HandshakePCAP:    new(string),
			}

			msg, err := stream.Recv()
			if err != nil {
				log.Panic(err.Error())
			}

			start := false
			for _, task := range msg.GetTasks() {
				// Identify an assigned task, it takes the first one if many are present
				if task.GetClientUuid() == clientUUID && task.GetStartCracking() {
					*handshake.HandshakePCAP = task.GetHashcatPcap()
					*handshake.ClientUUID = task.GetClientUuid()
					*handshake.HashcatOptions = task.GetHashcatOptions()
					handshake.UUID = task.GetHandshakeUuid()
					handshake.UserUUID = task.GetUserId()
					start = true
				}
			}

			if start {
				break
			}

			<-ticker.C
		}
		log.Println("[CLIENT] Task identified...")
		log.Println("[CLIENT] Decoding PCAP...")
		bytes, err := utils.StringBase64DataToBinary(*handshake.HandshakePCAP)
		if err != nil {
			log.Errorf("[CLIENT] Failed to decode PCAP %s", err.Error())
			continue
		}

		log.Println("[CLIENT] Saving pcap... ")
		pcapGenerated, err := utils.CreateMD5RandomFile(constants.TempPCAPStorage, constants.PCAPExtension, bytes)
		if err != nil {
			log.Errorf("[CLIENT] Failed to create temp PCAP file %s", err.Error())
			continue
		}

		randomHashcatFileName := filepath.Join(constants.TempHashcatFileDir, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(20))))+constants.HashcatExtension)
		err = hcxtools.ConvertPCAPToHashcatFormat(pcapGenerated, randomHashcatFileName)
		if err != nil {
			log.Errorf("[CLIENT] Failed to convert PCAP file using hcxtools %s", err.Error())
			continue
		}

		log.Println("[CLIENT] Checking generated conversion...")
		exists, err := utils.DirOrFileExists(randomHashcatFileName)

		if err != nil || !exists {
			log.Errorf("[CLIENT] Failed to verify generated hashcat file existence, maybe hcxtools failed...")
			continue
		}

		log.Println("[CLIENT] Running hashcat...")

		crackedHashes := map[string]*string{}

		hashcat, err := gocat.New(gocatOptions, goatCallback(crackedHashes))

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// -a 3 -m 22000 --potfile-disable --logfile-disable FILE_TO_CRACK test12?d?d
		replaced := strings.ReplaceAll(*handshake.HashcatOptions, "FILE_TO_CRACK", randomHashcatFileName)
		err = hashcat.RunJob(strings.Split(replaced, " ")...)

		if err != nil {
			log.Errorf("[CLIENT] Error on gocat command " + err.Error())
			continue
		}

		hashcat.Free()
		log.Println("[CLIENT] Finished hashcat.")

		var status = constants.ExhaustedStatus
		var result = ""
		for _, value := range crackedHashes {
			if value != nil {
				status = constants.CrackStatus
				result += *value
			}
		}
		log.Println(crackedHashes)

		finalize := &pb.ClientTaskMessageFromClient{
			Jwt:              *client.Credentials.JWT,
			HashcatLogs:      "TODO",
			CrackedHandshake: result,
			Status:           status,
			HandshakeUuid:    handshake.UUID,
			ClientUuid:       clientUUID,
		}

		err = stream.Send(finalize)
		newTicker := time.NewTicker(10 * time.Second)
		for err != nil {
			log.Println("[CLIENT] Handshake has been cracked but cannot update status to server, retrying in seconds")
			<-newTicker.C
			err = stream.Send(finalize)
		}
		<-newTicker.C
	}
}
