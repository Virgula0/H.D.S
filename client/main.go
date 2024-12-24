package main

import (
	"context"
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
var stateUpdateCh = make(chan gui.StateUpdate, 1)

// gocatCallback handles events from gocat and sends updates to the server stream.
func gocatCallback(
	resultsmap map[string]*string,
	stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer],
	msg *pb.ClientTaskMessageFromClient,
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

		// Send updated logs/stats to the server
		msg.HashcatLogs = logs.String()
		if err := stream.Send(msg); err != nil {
			return
		}
	}
}

// guiLogs periodically sends GUI updates until the context is cancelled.
func guiLogs(otherInfos map[string]string, ctx context.Context) {
	for {
		stateUpdateCh <- gui.StateUpdate{
			StatusLabel:   otherInfos[constants.HashcatStatus],
			IsConnected:   true,
			PCAPFile:      otherInfos[constants.PCAPFile],
			HashcatFile:   otherInfos[constants.HashcatFile],
			HashcatStatus: constants.CrackStatus,
			LogContent:    logs.String(),
		}
		if ctx.Err() != nil {
			log.Warn("Log goroutine killed")
			return
		}
	}
}

func main() {
	_, err := environment.InitEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	client := grpcclient.InitClient()
	defer client.ClientCloser()

	go gui.RunGUI(stateUpdateCh)

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
		log.Println("[CLIENT] Listening for tasks...")
		msg, err := stream.Recv()
		if err != nil {
			log.Errorf("[CLIENT] Closed connection: %s", err.Error())
			continue
		}

		logs.Reset() // Clear log buffer for each new task

		var handshake entities.Handshake
		handshake = entities.Handshake{
			ClientUUID:       new(string),
			CrackedDate:      new(string),
			HashcatOptions:   new(string),
			HashcatLogs:      new(string),
			CrackedHandshake: new(string),
			HandshakePCAP:    new(string),
		}

		for _, task := range msg.GetTasks() {
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
			logErrorAndSend(stream, client, handshake, constants.ErrorStatus, err.Error())
			continue
		}

		log.Println("[CLIENT] Saving pcap...")
		pcapGenerated, err := utils.CreateMD5RandomFile(constants.TempPCAPStorage, constants.PCAPExtension, bytes)
		if err != nil {
			logErrorAndSend(stream, client, handshake, constants.ErrorStatus, err.Error())
			continue
		}

		randomHashcatFileName := filepath.Join(
			constants.TempHashcatFileDir,
			fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(20))))+constants.HashcatExtension,
		)
		err = hcxtools.ConvertPCAPToHashcatFormat(pcapGenerated, randomHashcatFileName)
		if err != nil {
			logErrorAndSend(stream, client, handshake, constants.ErrorStatus, err.Error())
			continue
		}

		log.Println("[CLIENT] Checking generated conversion...")
		exists, err := utils.DirOrFileExists(randomHashcatFileName)
		if err != nil || !exists {
			logErrorAndSend(stream, client, handshake, constants.ErrorStatus, "hcxtools failed or file missing")
			continue
		}

		log.Println("[CLIENT] Running hashcat...")
		crackedHashes := map[string]*string{}

		msgToServer := &pb.ClientTaskMessageFromClient{
			Jwt:              *client.Credentials.JWT,
			HashcatLogs:      "",
			CrackedHandshake: "",
			Status:           constants.WorkingStatus,
			HandshakeUuid:    handshake.UUID,
			ClientUuid:       clientUUID,
			HashcatOptions:   *handshake.HashcatOptions,
		}

		hashcat, err := gocat.New(gocatOptions, gocatCallback(crackedHashes, stream, msgToServer))
		if err != nil {
			logErrorAndSend(stream, client, handshake, constants.ErrorStatus, err.Error())
			continue
		}

		logContext, killLogGoRoutine := context.WithCancel(context.Background())
		go guiLogs(map[string]string{
			constants.HashcatFile:   randomHashcatFileName,
			constants.PCAPFile:      pcapGenerated,
			constants.HashcatStatus: constants.WorkingStatus,
		}, logContext)

		replaced := strings.ReplaceAll(*handshake.HashcatOptions, "FILE_TO_CRACK", randomHashcatFileName)
		err = hashcat.RunJob(strings.Split(replaced, " ")...)
		var result, status string
		if err != nil {
			status = constants.ErrorStatus
		}

		hashcat.Free()
		log.Println("[CLIENT] Finished hashcat.")

		for _, value := range crackedHashes {
			if value != nil {
				status = constants.CrackStatus
				result += *value
			}
		}
		if len(crackedHashes) == 0 && status != constants.ErrorStatus {
			status = constants.ExhaustedStatus
		}

		killLogGoRoutine()

		finalize := &pb.ClientTaskMessageFromClient{
			Jwt:              *client.Credentials.JWT,
			HashcatLogs:      logs.String(),
			CrackedHandshake: result,
			Status:           status,
			HandshakeUuid:    handshake.UUID,
			ClientUuid:       clientUUID,
			HashcatOptions:   *handshake.HashcatOptions,
		}

		newTicker := time.NewTicker(30 * time.Second)
		for err := stream.Send(finalize); err != nil; err = stream.Send(finalize) {
			log.Println("[CLIENT] Failed to send final status, retrying in 30s...")
			<-newTicker.C
		}
	}
}

// logErrorAndSend is a helper that updates the logs with an error message and sends a failure status to the server.
func logErrorAndSend(
	stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer],
	client *grpcclient.Client,
	handshake entities.Handshake,
	status, errMsg string,
) {
	log.Errorf("[CLIENT] %s", errMsg)
	finalize := &pb.ClientTaskMessageFromClient{
		Jwt:            *client.Credentials.JWT,
		HashcatLogs:    logs.String(),
		Status:         status,
		HandshakeUuid:  handshake.UUID,
		ClientUuid:     *handshake.ClientUUID,
		HashcatOptions: *handshake.HashcatOptions,
	}
	_ = stream.Send(finalize)
}
