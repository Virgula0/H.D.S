package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/encryption"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
func InitClient() (*Client, error) {
	var conn *grpc.ClientConn
	var err error
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, //#nosec:G402 // use unsecure connection for first client installation or if security is disabled
		ServerName:         fmt.Sprintf("UnsecureConn-%s", utils.GenerateToken(32)),
	})

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

// ReloadConnectionWithCred will create a new connection with the updated credentials.
func (c *Client) ReloadConnectionWithCred(caCertPEM, clientKeyPEM, clientCertPEM []byte) error {

	log.Warn("[CLIENT] Reloading connection with TLS enabled")
	if err := c.client.Close(); err != nil { //close current connection
		return err
	}

	// Load credentials from in-memory bytes
	creds, err := encryption.LoadTLSCredentials(caCertPEM, clientKeyPEM, clientCertPEM)

	if err != nil {
		return fmt.Errorf("failed to load tls credentials %s %v", constants.GrpcURL, err)
	}

	// Create a new connection with the updated credentials
	conn, err := grpc.NewClient(constants.GrpcURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to reconnect with new credentials: %v", err)
	}

	log.Infof("[CLIENT] Connection instance re-created for %s", constants.GrpcURL)

	c.client = conn

	return nil
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
