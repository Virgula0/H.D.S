package mygocat

import (
	"crypto/md5" // #nosec G501
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/customerrors"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/internal/hcxtools"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	"github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

type TaskHandler struct {
	*Gocat
}

var firstLogTime = true

// ListenForHashcatTasks listens on the HashcatChat stream for tasks and processes them.
func (t *TaskHandler) ListenForHashcatTasks() error {
	log.Info("[CLIENT] Listening for tasks...")

	gui.StateUpdateCh <- &gui.StateUpdate{
		GRPCConnected: "Connected",
		StatusLabel:   "Listening for new tasks...",
		LogContent: func(logs string) string {
			if !firstLogTime {
				if strings.HasSuffix(logs, "\n") {
					return strings.Repeat("-", 200) + "\n"
				}
				return "\n" + strings.Repeat("-", 200) + "\n"
			}
			firstLogTime = false
			return ""
		}(grpcclient.ReadLogs()),
	}

	msg, err := t.Stream.Recv()
	if err != nil {
		return err
	}

	// Reset logs for each new batch of tasks
	grpcclient.ResetLogs()

	// Identify the handshake to be processed
	handshake, foundTask := t.identifyTask(msg.GetTasks())
	if !foundTask {
		// If no relevant task is found, simply return with no error; continue to wait for tasks
		log.Info("[CLIENT] No relevant tasks found. Waiting for next message...")
		return nil
	}

	log.Println("[CLIENT] Task identified...")
	err = t.processHandshakeTask(handshake)
	if err != nil {
		// update graphics with error
		gui.StateUpdateCh <- &gui.StateUpdate{
			HashcatStatus: constants.ErrorStatus,
			LogContent:    err.Error(),
		}

		return t.Client.LogErrorAndSend(t.Stream, handshake, constants.ErrorStatus, err.Error())
	}
	return nil
}

// identifyTask looks for a task matching the current client and returns a handshake struct if found.
//
//nolint:gocritic // false positive on nested reducing
func (t *TaskHandler) identifyTask(tasks []*hds.ClientTask) (*entities.Handshake, bool) {
	var handshake = &entities.Handshake{
		Status:           constants.PendingStatus,
		ClientUUID:       new(string),
		CrackedDate:      new(string),
		HashcatOptions:   new(string),
		HashcatLogs:      new(string),
		CrackedHandshake: new(string),
		HandshakePCAP:    new(string),
	}

	for _, task := range tasks {
		if task.GetClientUuid() == t.Client.EntityClient.ClientUUID && task.GetStartCracking() {
			*handshake.HandshakePCAP = task.GetHashcatPcap()
			*handshake.ClientUUID = task.GetClientUuid()
			*handshake.HashcatOptions = task.GetHashcatOptions()
			handshake.UUID = task.GetHandshakeUuid()
			handshake.UserUUID = task.GetUserId()
			handshake.SSID = task.GetSSID()
			handshake.BSSID = task.GetBSSID()
			return handshake, true
		}
	}
	return nil, false
}

// retrySendFinalStatus attempts to send the final status message to the server,
// retrying if there's a transient failure.
func (t *TaskHandler) retrySendFinalStatus(finalMsg *hds.ClientTaskMessageFromClient) error {
	tt := 30 * time.Second
	ticker := time.NewTicker(tt)
	defer ticker.Stop()

	gui.StateUpdateCh <- &gui.StateUpdate{
		HashcatStatus: finalMsg.GetStatus(),
		LogContent: func() string {
			// Check if print logs to GUI again or not
			if finalMsg.GetStatus() == constants.ErrorStatus {
				return finalMsg.GetHashcatLogs()
			}
			return ""
		}(),
	}

	for {
		if err := t.Stream.Send(finalMsg); err != nil {
			log.Errorf("%v %v", customerrors.ErrFinalSending.Error(), tt)
			<-ticker.C
			continue
		}
		return nil
	}
}

// ProcessHandshakeTask handles the entire process of decoding the PCAP, converting it,
// running Hashcat, and sending final status updates back to the server.
func (t *TaskHandler) processHandshakeTask(handshake *entities.Handshake) error {

	log.Println("[CLIENT] Decoding coming bytes...")
	data, err := utils.StringBase64DataToBinary(*handshake.HandshakePCAP)
	if err != nil {
		return err
	}

	log.Println("[CLIENT] Saving bytes...")
	// Generate a random filename for the Hashcat-ready file
	hashcatFilePath := filepath.Join(
		constants.TempHashcatFileDir,
		fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(20))))+constants.HashcatExtension, // #nosec G401
	)

	switch {
	case handshake.BSSID != "" && handshake.SSID != "":
		// If here, it means that input does not come from FE so we can threaten it has a handshake
		log.Println("[CLIENT] Converting pcap...")

		pcapFilePath, errFile := utils.CreateMD5RandomFile(constants.TempPCAPStorage, constants.PCAPExtension, data)

		if errFile != nil {
			return errFile
		}

		// Convert PCAP to Hashcat format, it actually created the hashcatFilePath
		if errConversion := hcxtools.ConvertPCAPToHashcatFormat(pcapFilePath, hashcatFilePath); errConversion != nil {
			return errConversion
		}

		// Ensure the conversion succeeded and file exists
		fileExists, errFile := utils.DirOrFileExists(hashcatFilePath)
		if errFile != nil {
			return errFile
		}

		if !fileExists {
			return customerrors.ErrHcxToolsNotFound
		}

		// Start cracking GUI info update
		gui.StateUpdateCh <- &gui.StateUpdate{
			PCAPFile: pcapFilePath,
		}

	default:
		// Else we do not need conversion, dump the file normally
		errCreateFile := utils.CreateFileWithBytes(hashcatFilePath, data)
		if errCreateFile != nil {
			return errCreateFile
		}
	}

	log.Println("[CLIENT] Running hashcat...")
	msgToServer := &hds.ClientTaskMessageFromClient{
		Jwt:            *t.Client.Credentials.JWT,
		Status:         constants.WorkingStatus,
		HandshakeUuid:  handshake.UUID,
		ClientUuid:     t.Client.EntityClient.ClientUUID,
		HashcatOptions: *handshake.HashcatOptions,
	}

	// Run the actual Hashcat operation
	finalStatusMsg, err := t.RunGoCat(
		msgToServer,
		hashcatFilePath,
		handshake,
	)

	if err != nil {
		log.Errorf("[CLIENT] Error in RunGoCat: %s", err.Error())
		return err
	}

	// Retry sending final status if needed
	return t.retrySendFinalStatus(finalStatusMsg)
}
