package raspberrypi

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/enums"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/entities"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var lastOperationMutex sync.Mutex

type TCPHandler interface {
	RunTCPServer()
}

type TCPCreateRaspberryPIRequest struct {
	Handshakes    []*entities.Handshake
	Jwt           string `validate:"required,jwt"`
	MachineID     string `validate:"required,len=32"`
	EncryptionKey string `validate:"omitempty,len=64"`
}

var statusACK = enums.ACK.String()
var statusErrorACK = enums.FAIL.String()

// RunTCPServer Start TCP server
func (wr *TCPServer) RunTCPServer() error {
	log.Printf("[TCP/IP Server] TCP/IP server running on %s", wr.l.Addr())

	for {
		client, err := wr.l.Accept()
		if err != nil {
			log.Errorf("[TCP/IP] Error accepting connection: %s", err.Error())
			continue
		}

		go wr.handleClientConnection(client)
	}
}

func (wr *TCPServer) timeOutClientManager(processTimeoutRequestErrChann chan error, lastOperation *time.Time) {
	defer func() {
		if _, closed := <-processTimeoutRequestErrChann; closed {
			return
		}
		close(processTimeoutRequestErrChann)
	}()

	lastOperationMutex.Lock()
	defer lastOperationMutex.Unlock()

	for {
		lastOperationMutex.Lock()
		if time.Since(*lastOperation) > wr.timeout {
			processTimeoutRequestErrChann <- fmt.Errorf("client timed out")
			return
		}
		lastOperationMutex.Unlock()
	}
}

func (wr *TCPServer) processClientRequestManager(processClientRequestErrChann chan error, client net.Conn, lastOperation *time.Time) {
	defer func() {
		if _, closed := <-processClientRequestErrChann; closed {
			return
		}
		close(processClientRequestErrChann)
	}()

	// for loop needed for sending both handshake and login using a single connection
	for {
		if err := wr.processClientRequest(client); err != nil && !errors.Is(err, customErrors.ErrHandshakeAlreadyPresent) {
			processClientRequestErrChann <- err
			return
		}
		lastOperationMutex.Lock()
		*lastOperation = time.Now()
		lastOperationMutex.Unlock()
	}
}

// handleClientConnection accept request from client
func (wr *TCPServer) handleClientConnection(client net.Conn) {
	defer client.Close()
	processClientRequestErrChann := make(chan error, 1)
	processTimeoutRequestErrChann := make(chan error, 1)
	lastOperation := time.Now()

	// routine for managing client time out
	go wr.timeOutClientManager(processTimeoutRequestErrChann, &lastOperation)

	// Start processing the client request in a separate goroutine
	go wr.processClientRequestManager(processClientRequestErrChann, client, &lastOperation)

	select {
	case err := <-processTimeoutRequestErrChann:
		log.Errorf("[TCP/IP] Error: %s", err.Error())
	case err := <-processClientRequestErrChann:
		log.Errorf("[TCP/IP] Error processing client request: %s", err.Error())
		return
	}
}

func (wr *TCPServer) sendACKToTheClient(client net.Conn) error {
	if _, err := client.Write([]byte(statusACK)); err != nil {
		return err
	}
	/*
		for some reason, it can happen that ACKs are read after error messages even if they're sent first. this may need further investigations
		The Nagle's Algorithm in TCP stack is not the problem, as by default is disabled in go
		there is no way to avoid this sleep, I tried everything, mutex, channels and so on...
	*/
	time.Sleep(wr.sleepTime)
	return nil
}

func (wr *TCPServer) sendACKFailedToTheClient(client net.Conn) error {
	if _, err := client.Write([]byte(statusErrorACK)); err != nil {
		return err
	}
	/*
		for some reason, it can happen that ACKs are read after error messages even if they're sent first. this may need further investigations
		The Nagle's Algorithm in TCP stack is not the problem, as by default is disabled in go
		there is no way to avoid this sleep, I tried everything, mutex, channels and so on...
	*/
	time.Sleep(wr.sleepTime)
	return nil
}

func (wr *TCPServer) handshake(client net.Conn) error {
	reader := bufio.NewReader(client)

	// Step 1: Read message size
	messageSize, errReadMsg := wr.readMessageSize(reader)
	if errReadMsg != nil {
		wr.writeErrorToClient(client, "Invalid message size")
		return errReadMsg
	}

	// Step 2: Send ACK to the client for the length
	if errFirstAckClient := wr.sendACKToTheClient(client); errFirstAckClient != nil {
		return errFirstAckClient
	}

	// Step 3: Read the actual message content
	buffer, err := wr.readMessageContent(reader, messageSize)
	if err != nil {
		wr.writeErrorToClient(client, "Error reading message content")
		return err
	}

	// Step 4: Send ACK to the client for the message
	if errSecondAckClient := wr.sendACKToTheClient(client); errSecondAckClient != nil {
		return errSecondAckClient
	}

	log.Printf("[TCP/IP] Received message: %s", string(buffer))

	// Step 5: Process the message type
	return wr.processHandshakeMessage(buffer, client)
}

func (wr *TCPServer) login(client net.Conn) error {
	reader := bufio.NewReader(client)
	// 1. Read message size
	messageSize, err := wr.readMessageSize(reader)
	if err != nil {
		wr.writeErrorToClient(client, "Invalid message size")
		return err
	}

	// Step 2: Send ACK to the client for the length
	if errFirstAckClient := wr.sendACKToTheClient(client); errFirstAckClient != nil {
		return errFirstAckClient
	}

	// Step 3: Read the actual message content
	buffer, err := wr.readMessageContent(reader, messageSize)
	if err != nil {
		wr.writeErrorToClient(client, "Error reading message content")
		return err
	}

	// Step 4: Send ACK to the client for the message
	if errSecondAckClient := wr.sendACKToTheClient(client); errSecondAckClient != nil {
		return errSecondAckClient
	}

	// Step 5: process login
	return wr.processLoginMessage(buffer, client)
}

// processClientRequest parses the client request
func (wr *TCPServer) processClientRequest(client net.Conn) error {
	reader := bufio.NewReader(client)
	// 1. Read message size
	messageSize, err := wr.readMessageSize(reader)
	if err != nil {
		wr.writeErrorToClient(client, "Invalid message size")
		return err
	}

	// Step 3: Send ACK of the message length to the client
	if errFirstAckClient := wr.sendACKToTheClient(client); errFirstAckClient != nil {
		return errFirstAckClient
	}

	// Step 4: Read the actual message content
	buffer, err := wr.readMessageContent(reader, messageSize)
	if err != nil {
		wr.writeErrorToClient(client, "Error reading message content")
		return err
	}

	// Step 3. Process login or command action
	switch string(buffer) {
	case enums.LOGIN.String():
		// Step 4: Send ACK of the message length to the client
		log.Infof("[TCP/IP] Received login message")
		if errSecondAckClient := wr.sendACKToTheClient(client); errSecondAckClient != nil {
			return errSecondAckClient
		}
		return wr.login(client)
	case enums.HANDSHAKE.String():
		// Step 4: Send ACK of the message length to the client
		log.Infof("[TCP/IP] Received handshake message")
		if errSecondAckClient := wr.sendACKToTheClient(client); errSecondAckClient != nil {
			return errSecondAckClient
		}
		return wr.handshake(client)
	default:
		// Step 4: Send ACK FAIL of command to the client
		log.Errorf("[TCP/IP] Received invalid command request: %s", string(buffer))
		if errSecondAckFailClient := wr.sendACKFailedToTheClient(client); errSecondAckFailClient != nil {
			return errSecondAckFailClient
		}
		return err
	}
}
