package grpcclient

import (
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"strings"
	"sync"
)

var logsMu sync.Mutex
var mutex sync.RWMutex
var logsSB = &strings.Builder{}

func AppendLog(s string) {
	logsMu.Lock()
	logsSB.WriteString(s)
	logsMu.Unlock()
}

func ReadLogs() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return logsSB.String()
}

// ResetLogs safely clears the content of the builder.
func ResetLogs() {
	logsMu.Lock()
	logsSB.Reset()
	logsMu.Unlock()
}

/*
Test

healthcheck to check if server is up and running
*/
func (c *Client) Test() (*pb.HelloResponse, error) {
	return c.PBInstance.Test(c.ClientContext, &pb.HelloRequest{
		Name: utils.GenerateToken(3),
	})
}

/*
HashcatChat

Initialize bidirectional stream channel for async client/server communication
*/
func (c *Client) HashcatChat() (grpc.BidiStreamingClient[pb.ClientTaskMessageFromClient, pb.ClientTaskMessageFromServer], error) {
	return c.PBInstance.HashcatTaskChat(c.ClientContext)
}

/*
GetClientInfo

calls the gRPC method for retrieving info. If the client exists server side it will be no registered.
*/
func (c *Client) GetClientInfo(name, machineID string) (*pb.GetClientInfoResponse, error) {
	return c.PBInstance.GetClientInfo(c.ClientContext, &pb.GetClientInfoRequest{
		Jwt:       *c.Credentials.JWT,
		MachineId: machineID,
		Name:      name,
	})
}

/*
Authenticate

Check if client can is authorized by the user
*/
func (c *Client) Authenticate(username, password string) (*pb.UniformResponse, error) {
	return c.PBInstance.Login(c.ClientContext, &pb.AuthRequest{
		Username: username,
		Password: password,
	})
}

// LogErrorAndSend is a helper that updates the logs with an error message and sends a failure status to the server.
func (c *Client) LogErrorAndSend(
	stream grpc.BidiStreamingClient[pb.ClientTaskMessageFromClient, pb.ClientTaskMessageFromServer],
	handshake *entities.Handshake,
	status, errMsg string,
) error {

	if ReadLogs() == "" {
		AppendLog(errMsg) // if here, error occurred outside gocat logger
	}

	log.Errorf("[CLIENT] %s", errMsg)
	finalize := &pb.ClientTaskMessageFromClient{
		Jwt:            *c.Credentials.JWT,
		HashcatLogs:    ReadLogs(),
		Status:         status,
		HandshakeUuid:  handshake.UUID,
		ClientUuid:     *handshake.ClientUUID,
		HashcatOptions: *handshake.HashcatOptions,
	}

	return stream.Send(finalize)
}
