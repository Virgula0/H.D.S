package testsuite

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"time"
)

type TCPServerSuite struct {
	suite.Suite
	Service      *restapi.ServiceHandler // contains Usecase as well for mocking
	DatabaseUser *infrastructure.Database
	DatabaseCert *infrastructure.Database
}

func (s *TCPServerSuite) SetupSuite() {
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
		service.InitRoutes(gorillaMux)
		restErr := srv.ListenAndServe()
		s.Require().NoError(restErr)
		log.Printf("[REST-API] Server running on %s:%s", constants.ServerHost, constants.ServerPort)
	}()

	s.startServer(&service)
}

func (s *TCPServerSuite) startServer(service *restapi.ServiceHandler) {
	server, err := raspberrypi.NewTCPServer(service, constants.TCPAddress, constants.TCPPort)
	s.Require().NoError(err)

	go func() {
		err = server.RunTCPServer()
		s.Require().NoError(err)
	}()

	// Init client
	time.Sleep(3 * time.Second) // give the time to start the server
}

func (s *TCPServerSuite) Client() net.Conn {
	conn, err := net.Dial("tcp", constants.TCPAddress+":"+constants.TCPPort)
	s.Require().NoError(err)

	return conn
}
