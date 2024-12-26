package grpcclient

import (
	"context"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type LoginInfo struct {
	JWT  *string
	Auth *entities.AuthRequest
}

type Client struct {
	client        *grpc.ClientConn
	PBInstance    pb.HDSTemplateServiceClient
	ClientContext context.Context
	ClientCloser  context.CancelFunc

	Credentials *LoginInfo
}

func InitClient() (*Client, error) {
	ticker := time.NewTicker(time.Second * 5)
	var conn *grpc.ClientConn
	var err error
	for {
		conn, err = grpc.NewClient(constants.GrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err == nil {
			log.Infof("[CLIENT] Connected with: %s", constants.GrpcURL)
			break
		}

		log.Errorf("[CLIENT] Error while attempting to connect to grpc server. Re-attempting in 5 seconds.")
		<-ticker.C
	}

	duration, err := time.ParseDuration(constants.GrpcTimeout)
	if err != nil {
		return nil, err
	}

	clientContext, cancel := context.WithTimeout(context.Background(), time.Duration(duration.Seconds())*time.Hour)

	return &Client{
		client:        conn,
		PBInstance:    pb.NewHDSTemplateServiceClient(conn),
		ClientContext: clientContext,
		ClientCloser:  cancel,

		Credentials: &LoginInfo{
			JWT:  new(string),
			Auth: new(entities.AuthRequest),
		},
	}, nil
}

func (c *Client) Authenticator() {
	ticker := time.NewTicker(1 * time.Hour) // every hour
	for {
		<-ticker.C
		// Every hour re-auth and re-update JWT
		if resp, err := c.Authenticate(c.Credentials.Auth.Username, c.Credentials.Auth.Password); err == nil {
			*c.Credentials.JWT = resp.GetDetails()
		} else {
			log.Fatal(err)
		}
	}
}
