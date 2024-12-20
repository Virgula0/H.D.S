package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	handlers "github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
)

func runService(m *mux.Router, database *infrastructure.Database) (*handlers.ServiceHandler, error) {
	ms, err := handlers.NewServiceHandler(database)
	if err != nil {
		return nil, fmt.Errorf("fail handlers.Handlers: %s", err.Error())
	}

	// Initialize routes on the default HTTP server mux
	ms.InitRoutes(m)
	return &ms, nil
}

func createServer(handler http.Handler, host, port string) *http.Server {
	return &http.Server{
		Addr:              host + ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// StartAsGRPC start the grpc_server server-grpc_server with the required business logic usecases
func startGRPC(service *handlers.ServiceHandler) error {
	grpc := grpcserver.New(grpcserver.NewServerContext(service.Usecase))

	timeout := constants.GrpcTimeout

	if constants.GrpcTimeoutParseError != nil {
		return constants.GrpcTimeoutParseError
	}

	err := grpc.Run(context.Background(), &grpcserver.Option{
		GrpcURL:         constants.GrpcURL,
		GrpcConnTimeout: timeout,
		Debug: func() bool {
			parsed, _ := strconv.ParseBool(constants.DebugEnabled)
			return parsed
		}(),
	})
	return err
}

func tcpServerInstance(service *handlers.ServiceHandler, host, port string) (*raspberrypi.TCPServer, error) {

	tcpInstance, err := raspberrypi.NewTCPServer(service, host, port)

	if err != nil {
		return nil, err
	}

	return tcpInstance, nil
}

func RunBackend() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// ---- SETUP REST API SERVER -> FOR FE AND RASPBERRY PI COMMUNICATIONS-----
	gorillaMux := mux.NewRouter()

	database, err := infrastructure.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	service, err := runService(gorillaMux, database)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	srv := createServer(gorillaMux, constants.ServerHost, constants.ServerPort)

	// Go Routine for the REST-API server
	go func() {
		go database.DBPinger()

		log.Printf("[REST-API] Server running on %s:%s", constants.ServerHost, constants.ServerPort)
		if restErr := srv.ListenAndServe(); restErr != nil && !errors.Is(restErr, http.ErrServerClosed) {
			log.Fatalf("listen: %v", restErr)
		}
	}()

	// ---- SETUP gRPC SERVER -> CLIENT COMMUNICATION -----

	go func() {
		err = startGRPC(service)
		if err != nil {
			log.Fatalf("Cannot start GRPC server! %s", err.Error())
		}
	}()

	// ---- SETUP TCP/IP SERVER -> RASPBERRY_PI COMMUNICATION -----

	tcpInstance, err := tcpServerInstance(service, constants.TCPAddress, constants.TCPPort)
	if err != nil {
		log.Fatalf("Cannot create TCP server instance! %s", err.Error())
	}

	err = tcpInstance.RunTCPServer()

	if err != nil {
		log.Fatalf("Cannot run TCP server! %s", err.Error())
	}
}
