//nolint:all // test
package main

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/internal/hcxtools"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	"github.com/Virgula0/progetto-dp/client/protobuf/hds"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
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

var logs strings.Builder

func gocatCallback(resultsmap map[string]*string,
	stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer],
	msg *pb.ClientTaskMessageFromClient,
	info *gui.ProcessWindowInfo,
	otherInfos map[string]string,
) gocat.EventCallback {
	return func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case gocat.LogPayload:
			if DebugTest {
				logs.WriteString(fmt.Sprintf("LOG [%s] %s\n", pl.Level, pl.Message))
				log.Infof("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.ActionPayload:
			if DebugTest {
				logs.WriteString(fmt.Sprintf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message))
				log.Infof("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.CrackedPayload:
			if DebugTest {
				logs.WriteString(fmt.Sprintf("CRACKED %s -> %s\n", pl.Hash, pl.Value))
				log.Infof("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
			if resultsmap != nil {
				resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
			}
		case gocat.FinalStatusPayload:
			if DebugTest {
				logs.WriteString(fmt.Sprintf("FINAL STATUS -> %v\n", pl.Status))
				log.Infof("FINAL STATUS -> %v\n", pl.Status)
			}
		case gocat.TaskInformationPayload:
			if DebugTest {
				logs.WriteString(fmt.Sprintf("TASK INFO -> %v\n", pl))
				log.Infof("TASK INFO -> %v\n", pl)
			}
		}

		// Send statistics to the server
		msg.HashcatLogs = logs.String()
		err := stream.Send(msg)
		if err != nil {
			return
		}

		// Update GUI
		info.Logs = logs.String()
		info.HashcatFile = otherInfos["hashcatFile"]
		info.HashcatStatus = otherInfos["status"]
		info.PCAPFile = otherInfos["pcapFile"]
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
	/*
		closed := gui.InitLoginWindow(client)

		if closed {
			os.Exit(1)
		}
	*/

	infoLogger := &gui.ProcessWindowInfo{}
	go func() {
		closed := infoLogger.InitProcessWindow()
		if closed {
			os.Exit(1)
		}
	}()

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

	for {
		var handshake entities.Handshake
		log.Println("[CLIENT] Listening for tasks...")
		infoLogger.IsConnected = true

		handshake = entities.Handshake{
			ClientUUID:       new(string),
			CrackedDate:      new(string),
			HashcatOptions:   new(string),
			HashcatLogs:      new(string),
			CrackedHandshake: new(string),
			HandshakePCAP:    new(string),
		}

		infoLogger.StatusLabel = "Waiting for tasks..."

		msg, err := stream.Recv()
		if err != nil {
			log.Errorf("[CLIENT] Closed connection %s", err.Error())
		}

		infoLogger.PCAPFile = ""
		infoLogger.HashcatFile = ""
		infoLogger.HashcatStatus = ""
		infoLogger.StatusLabel = "Task accepted " + handshake.UUID
		infoLogger.Logs = ""

		for _, task := range msg.GetTasks() {
			// Identify an assigned task, it takes the first one if many are present
			if task.GetClientUuid() == clientUUID && task.GetStartCracking() {
				*handshake.HandshakePCAP = task.GetHashcatPcap()
				*handshake.ClientUUID = task.GetClientUuid()
				*handshake.HashcatOptions = task.GetHashcatOptions()
				handshake.UUID = task.GetHandshakeUuid()
				handshake.UserUUID = task.GetUserId()
				break
			}
		}

		log.Println("[CLIENT] Task identified...")
		log.Println("[CLIENT] Decoding PCAP...")
		bytes, err := utils.StringBase64DataToBinary(*handshake.HandshakePCAP)
		if err != nil {
			log.Errorf("[CLIENT] Failed to decode PCAP %s", err.Error())
			finalize := &pb.ClientTaskMessageFromClient{
				Jwt:            *client.Credentials.JWT,
				HashcatLogs:    logs.String(),
				Status:         constants.ErrorStatus,
				HandshakeUuid:  handshake.UUID,
				ClientUuid:     clientUUID,
				HashcatOptions: *handshake.HashcatOptions,
			}
			err = stream.Send(finalize)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		log.Println("[CLIENT] Saving pcap... ")
		pcapGenerated, err := utils.CreateMD5RandomFile(constants.TempPCAPStorage, constants.PCAPExtension, bytes)
		if err != nil {
			log.Errorf("[CLIENT] Failed to create temp PCAP file %s", err.Error())
			finalize := &pb.ClientTaskMessageFromClient{
				Jwt:            *client.Credentials.JWT,
				HashcatLogs:    logs.String(),
				Status:         constants.ErrorStatus,
				HandshakeUuid:  handshake.UUID,
				ClientUuid:     clientUUID,
				HashcatOptions: *handshake.HashcatOptions,
			}
			err = stream.Send(finalize)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		randomHashcatFileName := filepath.Join(constants.TempHashcatFileDir, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(20))))+constants.HashcatExtension)
		err = hcxtools.ConvertPCAPToHashcatFormat(pcapGenerated, randomHashcatFileName)
		if err != nil {
			log.Errorf("[CLIENT] Failed to convert PCAP file using hcxtools %s", err.Error())
			finalize := &pb.ClientTaskMessageFromClient{
				Jwt:            *client.Credentials.JWT,
				HashcatLogs:    logs.String(),
				Status:         constants.ErrorStatus,
				HandshakeUuid:  handshake.UUID,
				ClientUuid:     clientUUID,
				HashcatOptions: *handshake.HashcatOptions,
			}
			err = stream.Send(finalize)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		log.Println("[CLIENT] Checking generated conversion...")
		exists, err := utils.DirOrFileExists(randomHashcatFileName)

		if err != nil || !exists {
			log.Error("[CLIENT] Failed to verify generated hashcat file existence, maybe hcxtools failed...")
			finalize := &pb.ClientTaskMessageFromClient{
				Jwt:            *client.Credentials.JWT,
				HashcatLogs:    logs.String(),
				Status:         constants.ErrorStatus,
				HandshakeUuid:  handshake.UUID,
				ClientUuid:     clientUUID,
				HashcatOptions: *handshake.HashcatOptions,
			}
			err = stream.Send(finalize)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		log.Println("[CLIENT] Running hashcat...")

		crackedHashes := map[string]*string{}

		hashcat, err := gocat.New(gocatOptions, gocatCallback(crackedHashes, stream, &pb.ClientTaskMessageFromClient{
			Jwt:              *client.Credentials.JWT,
			HashcatLogs:      "",
			CrackedHandshake: "",
			Status:           constants.WorkingStatus,
			HandshakeUuid:    handshake.UUID,
			ClientUuid:       clientUUID,
			HashcatOptions:   *handshake.HashcatOptions,
		},
			infoLogger,
			map[string]string{
				constants.HashcatFile:   randomHashcatFileName,
				constants.PCAPFile:      pcapGenerated,
				constants.HashcatStatus: constants.WorkingStatus,
			},
		))

		if err != nil {
			log.Errorf("Gocat init error %s", err.Error())
			finalize := &pb.ClientTaskMessageFromClient{
				Jwt:            *client.Credentials.JWT,
				HashcatLogs:    logs.String(),
				Status:         constants.ErrorStatus,
				HandshakeUuid:  handshake.UUID,
				ClientUuid:     clientUUID,
				HashcatOptions: *handshake.HashcatOptions,
			}
			err = stream.Send(finalize)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		// -a 3 -m 22000 --potfile-disable --logfile-disable FILE_TO_CRACK test12?d?d
		replaced := strings.ReplaceAll(*handshake.HashcatOptions, "FILE_TO_CRACK", randomHashcatFileName)
		err = hashcat.RunJob(strings.Split(replaced, " ")...)

		var result = ""
		var status = ""
		if err != nil {
			log.Errorf("[CLIENT] Error on gocat command %s", err.Error())
			status = constants.ErrorStatus
		}

		hashcat.Free()
		log.Println("[CLIENT] Finished hashcat.")

		for _, value := range crackedHashes {
			if value != nil {
				status = constants.CrackStatus
				infoLogger.StatusLabel = constants.CrackStatus
				result += *value
			} else {
				status = constants.ExhaustedStatus
				infoLogger.StatusLabel = constants.ExhaustedStatus

			}
		}
		log.Println(crackedHashes)

		finalize := &pb.ClientTaskMessageFromClient{
			Jwt:              *client.Credentials.JWT,
			HashcatLogs:      logs.String(),
			CrackedHandshake: result,
			Status:           status,
			HandshakeUuid:    handshake.UUID,
			ClientUuid:       clientUUID,
			HashcatOptions:   *handshake.HashcatOptions,
		}

		infoLogger.HashcatStatus = status
		logs.Reset()

		err = stream.Send(finalize)
		newTicker := time.NewTicker(30 * time.Second)
		for err != nil {
			log.Println("[CLIENT] Handshake has been cracked but cannot update status to server, retrying in seconds")
			<-newTicker.C
			err = stream.Send(finalize)
		}
	}
}
