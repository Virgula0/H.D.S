package grpcclient

import (
	"github.com/Virgula0/progetto-dp/client/protobuf/hds"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	"google.golang.org/grpc"
)

func (c *Client) HashcatChat() (grpc.BidiStreamingClient[hds.ClientTaskMessageFromClient, hds.ClientTaskMessageFromServer], error) {
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
