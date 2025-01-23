package raspberrypi

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	log "github.com/sirupsen/logrus"
	"net"
)

type TCPHandler interface {
	RunTCPServer()
}

type TCPCreateRaspberryPIRequest struct {
	Handshakes    []*entities.Handshake
	Jwt           string `validate:"required,jwt"`
	MachineID     string `validate:"required,len=32"`
	EncryptionKey string `validate:"omitempty,len=64"`
}

type ServerStatus int

const (
	ACK  = iota
	FAIL = 1
)

func (s ServerStatus) String() string {
	return [...]string{"ACK\n", "FAIL\n"}[s-1]
}

func (s ServerStatus) EnumIndex() int {
	return int(s)
}

// RunTCPServer Start TCP server
func (wr *TCPServer) RunTCPServer() error {
	log.Printf("[TCP/IP Server] TCP/IP server running on %s", wr.w.Addr())

	for {
		client, err := wr.w.Accept()
		if err != nil {
			log.Errorf("[TCP/IP] Error accepting connection: %s", err.Error())
			continue
		}

		go wr.handleClientConnection(client)
	}
}

// handleClientConnection accept request from client
func (wr *TCPServer) handleClientConnection(client net.Conn) {
	defer client.Close()

	// Set a timeout for the request handling
	ctx, cancel := context.WithTimeout(context.Background(), wr.timeout)
	defer cancel()

	done := make(chan error, 1)

	// let's do this for managing timeout connection
	go func() {
		done <- wr.processClientRequest(client) // process client request and prepare to read the next one
	}()

	select {
	case <-ctx.Done():
		log.Errorf("[TCP/IP] Request timed out for client: %s", client.RemoteAddr())
	case err := <-done:
		if err != nil {
			log.Errorf("[TCP/IP] Error processing request: %s", err.Error())
		}
	}
}

// processClientRequest parses the client request
func (wr *TCPServer) processClientRequest(client net.Conn) error {
	var status ServerStatus = FAIL
	reader := bufio.NewReader(client)

	// Step 1: Read message size
	messageSize, err := wr.readMessageSize(reader)
	if err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Invalid message size %s", status))
		return err
	}

	// Step 2: Read message content
	buffer, err := wr.readMessageContent(reader, messageSize)
	if err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Error reading message content %s", status))
		return err
	}

	// Step 3: Send ACK to the client
	status = ACK
	_, err = client.Write([]byte(status.String()))
	if err != nil {
		status = FAIL
		wr.writeErrorToClient(client, fmt.Sprintf("Error reading ACK %s", status))
		return err
	}

	log.Printf("[TCP/IP] Received message: %s", string(buffer))

	// Step 4: Process the message type
	return wr.processHandshakeMessage(buffer, client)
}
