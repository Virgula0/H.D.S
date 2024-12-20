package testsuite

import (
	"context"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	server "github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	usecaseHandler "github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
)

type GRPCTestSuite struct {
	suite.Suite
	Service  *restapi.ServiceHandler // contains Usecase as well
	Database *infrastructure.Database

	server        server.Server
	serverContext context.Context
	serverCloser  context.CancelFunc

	Client        pb.HDSTemplateServiceClient
	clientConn    *grpc.ClientConn
	clientContext context.Context
	clientCloser  context.CancelFunc
}

func (s *GRPCTestSuite) SetupSuite() {
	dbConn, err := infrastructure.NewDatabaseConnection()
	s.Require().NoError(err)
	s.Database = dbConn

	// Run rest api too
	service, err := restapi.NewServiceHandler(dbConn) // run seeds internally
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

	// server context
	srvCtx, srvCtxCancel := context.WithCancel(context.Background())
	s.Require().NotNil(srvCtx)
	s.Require().NotNil(srvCtxCancel)
	s.serverContext = srvCtx

	s.startServer(s.Service.Usecase)
}

func (s *GRPCTestSuite) startServer(usecase *usecaseHandler.Usecase) {

	// create server
	g := server.New(server.NewServerContext(usecase))
	s.server = *g

	errCh := make(chan error, 1)

	// run server in separate goroutine
	go func() {
		errCh <- g.Run(s.serverContext, &server.Option{
			GrpcURL:         constants.GrpcURL,
			GrpcConnTimeout: constants.GrpcTimeout,
			Debug:           false,
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
			conn, err := grpc.NewClient(constants.GrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
