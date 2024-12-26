package grpcclient

import (
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"strings"
)

var Logs = new(strings.Builder)

func (c *Client) HashcatChat() (grpc.BidiStreamingClient[pb.ClientTaskMessageFromClient, pb.ClientTaskMessageFromServer], error) {
	return c.PBInstance.HashcatTaskChat(c.ClientContext)
}

func (c *Client) GetClientInfo(name, machineID string) (*pb.GetClientInfoResponse, error) {
	return c.PBInstance.GetClientInfo(c.ClientContext, &pb.GetClientInfoRequest{
		Jwt:       *c.Credentials.JWT,
		MachineId: machineID,
		Name:      name,
	})
}

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
) {
	log.Errorf("[CLIENT] %s", errMsg)
	finalize := &pb.ClientTaskMessageFromClient{
		Jwt:            *c.Credentials.JWT,
		HashcatLogs:    Logs.String(),
		Status:         status,
		HandshakeUuid:  handshake.UUID,
		ClientUuid:     *handshake.ClientUUID,
		HashcatOptions: *handshake.HashcatOptions,
	}
	_ = stream.Send(finalize)
}
