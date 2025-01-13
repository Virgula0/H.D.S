package mygocat

import (
	"context"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
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
	stream grpc.BidiStreamingClient[pb.ClientTaskMessageFromClient, pb.ClientTaskMessageFromServer],
	msg *pb.ClientTaskMessageFromClient,
) gocat.EventCallback {

	return func(hc unsafe.Pointer, payload interface{}) {
		handlePayload(payload, resultsmap)

		// Update logs/stats and send them to the server
		msg.HashcatLogs = grpcclient.ReadLogs()
		if err := stream.Send(msg); err != nil {
			log.Errorf("Failed to send message to server: %v", err)
		}
	}
}

// ---------- PAYLOAD HANDLERS ----------

// handlePayload dispatches the payload to specific handlers based on its type.
func handlePayload(payload any, resultsmap map[string]*string) {
	switch pl := payload.(type) {
	case gocat.LogPayload:
		handleLogPayload(&pl)
	case gocat.ActionPayload:
		handleActionPayload(&pl)
	case gocat.CrackedPayload:
		handleCrackedPayload(&pl, resultsmap)
	case gocat.FinalStatusPayload:
		handleFinalStatusPayload(&pl)
	case gocat.TaskInformationPayload:
		handleTaskInformationPayload(&pl)
	case gocat.Status:
		handleTaskStatus(&pl)
	case gocat.DeviceStatus:
		handleDeviceStatus(&pl)
	case gocat.ErrCrackedPayload:
		handleErrCrackedPayload(&pl)
	case nil:
		return
	default:
		log.Warnf("Unhandled payload type: %v:%T ", pl, payload)
	}
}

// ---------- SPECIFIC HANDLER FUNCTIONS ----------

func handleLogPayload(pl *gocat.LogPayload) {
	if DebugTest {
		logMessage := fmt.Sprintf("LOG [%s] %s\n", pl.Level, pl.Message)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleActionPayload(pl *gocat.ActionPayload) {
	if DebugTest {
		logMessage := fmt.Sprintf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleCrackedPayload(pl *gocat.CrackedPayload, resultsmap map[string]*string) {
	if DebugTest {
		logMessage := fmt.Sprintf("CRACKED %s -> %s\n", pl.Hash, pl.Value)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
	if resultsmap != nil {
		resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
	}
}

func handleFinalStatusPayload(pl *gocat.FinalStatusPayload) {
	if DebugTest {
		logMessage := fmt.Sprintf("FINAL STATUS -> %v\n", pl.Status)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleTaskInformationPayload(pl *gocat.TaskInformationPayload) {
	if DebugTest {
		logMessage := fmt.Sprintf("TASK INFO -> %v\n", pl)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleTaskStatus(pl *gocat.Status) {
	if DebugTest {
		logMessage := fmt.Sprintf("CURRENT STATUS -> %v\n", pl)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleDeviceStatus(pl *gocat.DeviceStatus) {
	if DebugTest {
		logMessage := fmt.Sprintf("DEVICE STATUS -> %v\n", pl)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}

func handleErrCrackedPayload(pl *gocat.ErrCrackedPayload) {
	if DebugTest {
		logMessage := fmt.Sprintf("DEVICE STATUS -> %v\n", pl)
		grpcclient.AppendLog(logMessage)
		log.Info(logMessage)
	}
}
func RunGoCat(
	stream grpc.BidiStreamingClient[pb.ClientTaskMessageFromClient, pb.ClientTaskMessageFromServer],
	msgToServer *pb.ClientTaskMessageFromClient,
	randomHashcatFileName string,
	handshake *entities.Handshake,
	client *grpcclient.Client,
) (*pb.ClientTaskMessageFromClient, error) {

	logContext, killLogGoRoutine := context.WithCancel(context.Background())
	defer killLogGoRoutine()

	// Launch go-routine for updating logs dinamically
	go gui.GuiLogger(logContext, gui.StateUpdateCh)

	// Start cracking GUI info update
	gui.StateUpdateCh <- &gui.StateUpdate{
		StatusLabel:   handshake.UUID,
		HashcatFile:   randomHashcatFileName,
		HashcatStatus: constants.WorkingStatus,
	}

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

	replaced := strings.ReplaceAll(*handshake.HashcatOptions, constants.FileToCrackPlaceHolder, randomHashcatFileName)
	err = hashcat.RunJob(strings.Split(replaced, " ")...)
	var result, status string

	if err != nil {
		return &pb.ClientTaskMessageFromClient{
			Jwt:            *client.Credentials.JWT,
			HashcatLogs:    fmt.Sprintf("[%s] with command '%s'", err.Error(), replaced),
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
		HashcatLogs:      grpcclient.ReadLogs(),
		CrackedHandshake: result,
		Status:           status,
		HandshakeUuid:    handshake.UUID,
		ClientUuid:       *handshake.ClientUUID,
		HashcatOptions:   *handshake.HashcatOptions,
	}, nil
}
