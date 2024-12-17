package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/grpc_server"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	handlers "github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
)

var ServerHost = os.Getenv("BACKEND_HOST")
var ServerPort = os.Getenv("BACKEND_PORT")

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
	grpc := grpc_server.New(grpc_server.NewServerContext(service.Usecase))

	timeout := constants.GRPC_TIMEOUT

	if constants.GRPC_TIMEOUT_PARSE_ERROR != nil {
		return constants.GRPC_TIMEOUT_PARSE_ERROR
	}

	err := grpc.Run(context.Background(), &grpc_server.Option{
		GrpcURL:         constants.GRPC_URL,
		GrpcConnTimeout: timeout,
		Debug: func() bool {
			parsed, _ := strconv.ParseBool(constants.DebugEnabled)
			return parsed
		}(),
	})
	return err
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

	srv := createServer(gorillaMux, ServerHost, ServerPort)

	// Go Routine for the REST-API server
	go func() {
		go database.DBPinger()

		log.Printf("Server running on %s:%s", ServerHost, ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// ---- SETUP gRPC SERVER -> CLIENT COMMUNICATION -----

	err = startGRPC(service)

	if err != nil {
		log.Fatalf("Cannot start GRPC server! %s", err.Error())
	}

}
