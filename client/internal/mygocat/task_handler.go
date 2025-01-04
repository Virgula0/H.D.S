package mygocat

import (
	"crypto/md5" // #nosec G501
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/internal/hcxtools"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	"github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"path/filepath"
	"time"
)

// ListenForHashcatTasks listens on the HashcatChat stream for tasks and processes them.
func ListenForHashcatTasks(stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer], client *grpcclient.Client, clientUUID string) error {
	log.Println("[CLIENT] Listening for tasks...")

	gui.StateUpdateCh <- &gui.StateUpdate{
		GRPCConnected: "Connected",
		StatusLabel:   "Listening for new tasks...",
	}

	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	// Reset logs for each new batch of tasks
	grpcclient.ResetLogs()

	// Identify the handshake to be processed
	handshake, foundTask := identifyTask(msg.GetTasks(), clientUUID)
	if !foundTask {
		// If no relevant task is found, simply return with no error; continue to wait for tasks
		log.Println("[CLIENT] No relevant tasks found. Waiting for next message...")
		return nil
	}

	log.Println("[CLIENT] Task identified...")
	return processHandshakeTask(stream, client, clientUUID, handshake)
}

// identifyTask looks for a task matching the current client and returns a handshake struct if found.
//
//nolint:gocritic // false positive on nested reducing
func identifyTask(tasks []*hds.ClientTask, clientUUID string) (*entities.Handshake, bool) {
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
		if task.GetClientUuid() == clientUUID && task.GetStartCracking() {
			*handshake.HandshakePCAP = task.GetHashcatPcap()
			*handshake.ClientUUID = task.GetClientUuid()
			*handshake.HashcatOptions = task.GetHashcatOptions()
			handshake.UUID = task.GetHandshakeUuid()
			handshake.UserUUID = task.GetUserId()
			return handshake, true
		}
	}
	return nil, false
}

// retrySendFinalStatus attempts to send the final status message to the server,
// retrying if there's a transient failure.
func retrySendFinalStatus(stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer], finalMsg *hds.ClientTaskMessageFromClient) error {
	ticker := time.NewTicker(30 * time.Second)
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
		if err := stream.Send(finalMsg); err != nil {
			log.Println("[CLIENT] Failed to send final status, retrying in 30s...")
			<-ticker.C
			continue
		}
		return nil
	}
}

// ProcessHandshakeTask handles the entire process of decoding the PCAP, converting it,
// running Hashcat, and sending final status updates back to the server.
func processHandshakeTask(
	stream grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer],
	client *grpcclient.Client,
	clientUUID string,
	handshake *entities.Handshake,
) error {

	log.Println("[CLIENT] Decoding PCAP...")
	pcapData, err := utils.StringBase64DataToBinary(*handshake.HandshakePCAP)
	if err != nil {
		client.LogErrorAndSend(stream, handshake, constants.ErrorStatus, err.Error())
		return nil
	}

	log.Println("[CLIENT] Saving pcap...")
	pcapFilePath, err := utils.CreateMD5RandomFile(constants.TempPCAPStorage, constants.PCAPExtension, pcapData)
	if err != nil {
		client.LogErrorAndSend(stream, handshake, constants.ErrorStatus, err.Error())
		return nil
	}

	// Generate a random filename for the Hashcat-ready file
	hashcatFilePath := filepath.Join(
		constants.TempHashcatFileDir,
		fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(20))))+constants.HashcatExtension, // #nosec G401
	)

	// Convert PCAP to Hashcat format
	if err = hcxtools.ConvertPCAPToHashcatFormat(pcapFilePath, hashcatFilePath); err != nil {
		client.LogErrorAndSend(stream, handshake, constants.ErrorStatus, err.Error())
		return nil
	}

	// Ensure the conversion succeeded and file exists
	fileExists, err := utils.DirOrFileExists(hashcatFilePath)
	if err != nil {
		client.LogErrorAndSend(stream, handshake, constants.ErrorStatus, err.Error())
		return nil
	}

	if !fileExists {
		client.LogErrorAndSend(stream, handshake, constants.ErrorStatus, "conversion was not successful, hcxtools output file not found")
	}

	log.Println("[CLIENT] Running hashcat...")
	msgToServer := &hds.ClientTaskMessageFromClient{
		Jwt:            *client.Credentials.JWT,
		Status:         constants.WorkingStatus,
		HandshakeUuid:  handshake.UUID,
		ClientUuid:     clientUUID,
		HashcatOptions: *handshake.HashcatOptions,
	}

	// Run the actual Hashcat operation
	finalStatusMsg, err := RunGoCat(
		stream,
		msgToServer,
		hashcatFilePath,
		pcapFilePath,
		handshake,
		client,
	)
	if err != nil {
		log.Errorf("[CLIENT] Error in RunGoCat: %s", err.Error())
	}

	// Retry sending final status if needed
	return retrySendFinalStatus(stream, finalStatusMsg)
}
