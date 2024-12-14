package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/handlers"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/gorilla/mux"
)

var ServerHost = os.Getenv("BACKEND_HOST")
var ServerPort = os.Getenv("BACKEND_PORT")

const WipeTables = true

func runService(mux *mux.Router, database *infrastructure.Database) error {
	ms, err := handlers.NewServiceHandler(database, WipeTables)
	if err != nil {
		return fmt.Errorf("fail handlers.Handlers: %s", err.Error())
	}

	// Initialize routes on the default HTTP server mux
	ms.InitRoutes(mux)
	return nil
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

func main() {
	// ---- SETUP REST API SERVER -> FOR FE AND RASPBERRY PI COMMUNICATIONS-----
	runtime.GOMAXPROCS(runtime.NumCPU())

	gorillaMux := mux.NewRouter()

	database, err := infrastructure.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	go database.DBPinger()

	err = runService(gorillaMux, database)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	srv := createServer(gorillaMux, ServerHost, ServerPort)

	log.Printf("Server running on %s:%s", ServerHost, ServerPort)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}

	// ---- SETUP gRPC SERVER -> CLIENT COMMUNICATION -----

}
