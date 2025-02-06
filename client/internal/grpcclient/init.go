package grpcclient

import (
	"context"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/encryption"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

	Credentials  *LoginInfo
	EntityClient *entities.Client
}

/*
InitClient

Initialize gRPC client
*/
func InitClient(env *environment.Environment) (*Client, error) {
	var conn *grpc.ClientConn
	var err error

	// use mTLS or not
	creds, err := encryption.LoadTLSCredentials(env.Keys.CACert, env.Keys.ClientKey, env.Keys.ClientCert, !env.EmptyCerts())
	if err != nil {
		return nil, err
	}

	conn, err = grpc.NewClient(constants.GrpcURL, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Fatalf("[CLIENT] Cannot enstablish a connection with server %s %v", constants.GrpcURL, err)
	}

	log.Infof("[CLIENT] Connection instance created for %s", constants.GrpcURL)

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

/*
Authenticator

runs in background, and each hour using provided credentials, updates the JWT token required for performing operations
server side
*/
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
