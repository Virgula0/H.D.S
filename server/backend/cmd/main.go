package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver/encryption"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	handlers "github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
)

// runService initialize service infrastructure connecting to the database and saving the instance
func runService(m *mux.Router, dbUser, dbCerts *infrastructure.Database) (*handlers.ServiceHandler, error) {
	ms, err := handlers.NewServiceHandler(dbUser, dbCerts)
	if err != nil {
		return nil, fmt.Errorf("fail handlers.Handlers: %s", err.Error())
	}

	// Initialize routes on the default HTTP server mux
	ms.InitRoutes(m)
	return &ms, nil
}

// createServer initialize httpserver for restapi
func createServer(handler http.Handler, host, port string) *http.Server {
	return &http.Server{
		Addr:              host + ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func updateClientConfigRuntime(service *handlers.ServiceHandler, storage *encryption.ClientConfigStore) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		clients, _, err := service.Usecase.GetClientsInstalled()

		if err != nil {
			log.Errorf("fail to get clients: %s", err.Error())
		}

		for _, client := range clients {
			storage.UpdateClientConfig(client.ClientUUID, &encryption.ClientConfig{
				EncryptionEnabled: client.EnabledEncryption,
				ID:                client.ClientUUID,
			})
		}

		<-ticker.C
	}
}

// StartAsGRPC start the grpc_server server-grpc_server with the required business logic usecases
func startGRPC(service *handlers.ServiceHandler) error {
	grpc := grpcserver.New(grpcserver.NewServerContext(service.Usecase))

	timeout := constants.GrpcTimeout

	if constants.GrpcTimeoutParseError != nil {
		return constants.GrpcTimeoutParseError
	}

	caCert, caKey, serverCert, serverKey, certErr := service.Usecase.GetServerCerts()

	if certErr != nil {
		return certErr
	}

	storage := encryption.NewClientCertStore()

	// for each client update client config already existing from db
	// this is needed since we need to update client certs with new generated server keys
	clients, _, errInstalled := service.Usecase.GetClientsInstalled()
	if errInstalled != nil {
		return errInstalled
	}

	for _, client := range clients {
		// update in db, but only if the encryption was enabled
		if err := service.Usecase.UpdateCerts(client); err != nil {
			return err
		}
	}

	go updateClientConfigRuntime(service, storage)

	return grpc.Run(context.Background(), &grpcserver.Option{
		Debug:               constants.DebugEnabled,
		GrpcURL:             constants.GrpcURL,
		GrpcConnTimeout:     timeout,
		CACert:              caCert,
		CAKey:               caKey,
		ServerCert:          serverCert,
		ServerKey:           serverKey,
		ClientConfigStorage: storage,
	})
}

// tcpServerInstance initialize tcp server
func tcpServerInstance(service *handlers.ServiceHandler, host, port string) (*raspberrypi.TCPServer, error) {

	tcpInstance, err := raspberrypi.NewTCPServer(service, host, port)

	if err != nil {
		return nil, err
	}

	return tcpInstance, nil
}

func RunBackend() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	dbUser, err := infrastructure.NewDatabaseConnection(constants.DBUser, constants.DBPassword, constants.DBHost, constants.DBPort, constants.DBName)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	dbCerts, err := infrastructure.NewDatabaseConnection(constants.DBCertUser, constants.DBCertPass, constants.DBHost, constants.DBPort, constants.DBCert)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	// ---- SETUP REST API SERVER -> FOR FE COMMUNICATIONS-----
	gorillaMux := mux.NewRouter()

	service, err := runService(gorillaMux, dbUser, dbCerts)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	srv := createServer(gorillaMux, constants.ServerHost, constants.ServerPort)

	// Go Routine for the REST-API server
	go func() {
		go dbUser.StartDBPinger()
		log.Infof("[REST-API] Server running on %s:%s", constants.ServerHost, constants.ServerPort)
		if restErr := srv.ListenAndServe(); restErr != nil && !errors.Is(restErr, http.ErrServerClosed) {
			log.Fatalf("listen: %v", restErr)
		}
	}()

	// ---- SETUP gRPC SERVER -> CLIENT COMMUNICATION -----
	err = service.Usecase.CreateServerCerts() // create certs for mTLS scopes
	if err != nil {
		log.Fatalf("Cannot create mTLS certs! %s", err.Error())
	}

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

	go func() {
		err = tcpInstance.RunTCPServer()

		if err != nil {
			log.Fatalf("Cannot run TCP server! %s", err.Error())
		}
	}()
}
