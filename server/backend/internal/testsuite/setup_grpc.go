package testsuite

import (
	"context"
	"crypto/tls"
	"github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver/encryption"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	server "github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type GRPCTestSuite struct {
	suite.Suite
	Service      *restapi.ServiceHandler // contains Usecase as well
	DatabaseUser *infrastructure.Database
	DatabaseCert *infrastructure.Database

	server        server.Server
	serverContext context.Context
	serverCloser  context.CancelFunc

	Client        pb.HDSTemplateServiceClient
	clientConn    *grpc.ClientConn
	clientContext context.Context
	clientCloser  context.CancelFunc
}

func (s *GRPCTestSuite) SetupSuite() {
	dbConnUser, err := infrastructure.NewDatabaseConnection(constants.DBUser, constants.DBPassword, constants.DBHost, constants.DBPort, constants.DBName)
	s.Require().NoError(err)
	s.DatabaseUser = dbConnUser

	dbConnCerts, err := infrastructure.NewDatabaseConnection(constants.DBCertUser, constants.DBCertPass, constants.DBHost, constants.DBPort, constants.DBCert)
	s.Require().NoError(err)
	s.DatabaseCert = dbConnCerts

	// Run rest api too
	service, err := restapi.NewServiceHandler(dbConnUser, dbConnCerts) // run seeds internally
	s.Require().NoError(err)
	s.Service = &service

	gorillaMux := mux.NewRouter()

	srv := &http.Server{
		Addr:              constants.ServerHost + ":" + constants.ServerPort,
		Handler:           gorillaMux,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Run REST-API
	go func() {
		log.Printf("[REST-API] Server running on %s:%s", constants.ServerHost, constants.ServerPort)
		service.InitRoutes(gorillaMux)
		restErr := srv.ListenAndServe()
		s.Require().NoError(restErr)
	}()

	// create server certs, TODO: tests
	err = s.Service.Usecase.CreateServerCerts()
	s.Require().NoError(err)

	// server context
	srvCtx, srvCtxCancel := context.WithCancel(context.Background())
	s.Require().NotNil(srvCtx)
	s.Require().NotNil(srvCtxCancel)
	s.serverContext = srvCtx

	s.startServer()
}

func (s *GRPCTestSuite) startServer() {

	// create server
	g := server.New(server.NewServerContext(s.Service.Usecase))
	s.server = *g

	errCh := make(chan error, 1)

	caCert, caKey, serverCert, serverKey, err := s.Service.Usecase.GetServerCerts()
	s.Require().NoError(err)

	// run server in separate goroutine
	go func() {
		errCh <- g.Run(s.serverContext, &server.Option{
			Debug:               true,
			GrpcURL:             constants.GrpcURL,
			GrpcConnTimeout:     constants.GrpcTimeout,
			CACert:              caCert,
			CAKey:               caKey,
			ServerCert:          serverCert,
			ServerKey:           serverKey,
			ClientConfigStorage: encryption.NewClientCertStore(),
		})
	}()

	// start g client
	s.startClient(errCh)
}

func (s *GRPCTestSuite) startClient(errCh chan error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	// create client
	for {
		select {
		case <-time.After(time.Second):
			creds := credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})
			conn, err := grpc.NewClient(constants.GrpcURL, grpc.WithTransportCredentials(creds))
			if err == nil {
				// Connection established
				s.clientConn = conn
				c := pb.NewHDSTemplateServiceClient(conn)
				s.Require().NotNil(c)
				s.Client = c
				// create client context
				clientContext, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				s.Require().NotNil(clientContext)
				s.Require().NotNil(cancel)
				s.clientContext = clientContext
				s.clientCloser = cancel

				_, err = c.Test(ctx, &pb.HelloRequest{Name: "gaoff"})
				s.Require().NoError(err)
				return // exit SetupSuite()
			}
		case err := <-errCh:
			s.Require().NoError(err)
			return // exit SetupSuite()
		case <-ctx.Done():
			// If the server hasn't started in 10 seconds, stop trying to connect and shut it down
			s.serverCloser()
			s.Fail("Server didn't start within 10 seconds")
			return // exit SetupSuite()
		}
	}
}
