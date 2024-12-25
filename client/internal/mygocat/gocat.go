package mygocat

import (
	"context"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/protobuf/hds"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	"github.com/mandiant/gocat/v6"
	"github.com/mandiant/gocat/v6/hcargp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"strings"
	"unsafe"
)

const DebugTest = true

var gocatOptions = gocat.Options{
	SharedPath: "/usr/local/share/hashcat/OpenCL",
}

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
				grpcclient.Logs.WriteString(fmt.Sprintf("LOG [%s] %s\n", pl.Level, pl.Message))
				log.Infof("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.ActionPayload:
			if DebugTest {
				grpcclient.Logs.WriteString(fmt.Sprintf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message))
				log.Infof("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.CrackedPayload:
			if DebugTest {
				grpcclient.Logs.WriteString(fmt.Sprintf("CRACKED %s -> %s\n", pl.Hash, pl.Value))
				log.Infof("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
			if resultsmap != nil {
				resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
			}
		case gocat.FinalStatusPayload:
			if DebugTest {
				grpcclient.Logs.WriteString(fmt.Sprintf("FINAL STATUS -> %v\n", pl.Status))
				log.Infof("FINAL STATUS -> %v\n", pl.Status)
			}
		case gocat.TaskInformationPayload:
			if DebugTest {
				grpcclient.Logs.WriteString(fmt.Sprintf("TASK INFO -> %v\n", pl))
				log.Infof("TASK INFO -> %v\n", pl)
			}
		}

		// Send updated logs/stats to the server
		msg.HashcatLogs = grpcclient.Logs.String()
		if err := stream.Send(msg); err != nil {
			return
		}
	}
}

func RunGoCat(
	stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer],
	msgToServer *pb.ClientTaskMessageFromClient,
	randomHashcatFileName, pcapGenerated string,
	handshake *entities.Handshake,
	client *grpcclient.Client,
) (*pb.ClientTaskMessageFromClient, error) {

	crackedHashes := map[string]*string{}
	hashcat, err := gocat.New(gocatOptions, gocatCallback(crackedHashes, stream, msgToServer))
	defer hashcat.Free()

	if err != nil {
		return &pb.ClientTaskMessageFromClient{
			Jwt:            *client.Credentials.JWT,
			HashcatLogs:    err.Error(),
			Status:         constants.ErrorStatus,
			HandshakeUuid:  handshake.UUID,
			ClientUuid:     *handshake.ClientUUID,
			HashcatOptions: *handshake.HashcatOptions,
		}, err
	}

	logContext, killLogGoRoutine := context.WithCancel(context.Background())
	defer killLogGoRoutine()

	go gui.GuiLogger(map[string]string{
		constants.HashcatFile:   randomHashcatFileName,
		constants.PCAPFile:      pcapGenerated,
		constants.HashcatStatus: constants.WorkingStatus,
	}, logContext)

	replaced := strings.ReplaceAll(*handshake.HashcatOptions, constants.FileToCrackPlaceHolder, randomHashcatFileName)
	err = hashcat.RunJob(strings.Split(replaced, " ")...)
	var result, status string

	if err != nil {
		return &pb.ClientTaskMessageFromClient{
			Jwt:            *client.Credentials.JWT,
			HashcatLogs:    err.Error(),
			Status:         constants.ErrorStatus,
			HandshakeUuid:  handshake.UUID,
			ClientUuid:     *handshake.ClientUUID,
			HashcatOptions: *handshake.HashcatOptions,
		}, err
	}

	log.Println("[CLIENT] Finished hashcat.")

	for _, value := range crackedHashes {
		if value != nil {
			status = constants.CrackStatus
			result += *value
		}
	}

	if len(crackedHashes) == 0 {
		status = constants.ExhaustedStatus
	}

	msgToServer.Status = status
	msgToServer.CrackedHandshake = result

	return &pb.ClientTaskMessageFromClient{
		Jwt:              *client.Credentials.JWT,
		HashcatLogs:      grpcclient.Logs.String(),
		CrackedHandshake: result,
		Status:           status,
		HandshakeUuid:    handshake.UUID,
		ClientUuid:       *handshake.ClientUUID,
		HashcatOptions:   *handshake.HashcatOptions,
	}, nil
}
