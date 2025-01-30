// notlint: all
package grpcserver_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/testsuite"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type userClientRegistered struct {
	*entities.Client
	clientCert []byte
	clientKey  []byte
}

type GRPCServerTestSuite struct {
	testsuite.GRPCTestSuite
	UserFixture      *entities.User
	UserTokenFixture string

	UserClientRegistered   userClientRegistered
	UserClientUnregistered *entities.Client

	NormalUserFixture      *entities.User
	NormalUserTokenFixture string

	HandshakeValidID string
}

// Run All tests
func TestGRPCMethodCaller(t *testing.T) {
	suite.Run(t, new(GRPCServerTestSuite))
}

func (s *GRPCServerTestSuite) SetupSuite() {
	s.GRPCTestSuite.SetupSuite()
	s.Require().NotNil(seed.AdminSeed.User)
	s.UserFixture = seed.AdminSeed.User

	s.Require().NotNil(seed.NormalUserSeed.User)
	s.NormalUserFixture = seed.NormalUserSeed.User

	log.Error(s.UserFixture)

	// Create a client known
	s.UserClientRegistered = userClientRegistered{
		Client: &entities.Client{
			UserUUID:  s.UserFixture.UserUUID,
			Name:      "TEST",
			MachineID: fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))),
		},
	}

	clientID, err := s.Service.Usecase.CreateClient(s.UserFixture.UserUUID, s.UserClientRegistered.MachineID, s.UserClientRegistered.LatestIP, s.UserClientRegistered.Name)
	s.Require().NoError(err)
	// assign generated clientID
	s.UserClientRegistered.ClientUUID = clientID

	caCert, caKey, err := s.Service.Usecase.GetServerCerts()
	s.Require().NoError(err)

	// sign certs
	clientCert, clientKey, err := s.Service.Usecase.SignCert(caCert, caKey, clientID)
	s.Require().NoError(err)

	_, err = s.Service.Usecase.CreateCertForClient(clientID, clientCert, clientKey)
	s.Require().NoError(err)

	s.UserClientRegistered.clientCert = clientCert
	s.UserClientRegistered.clientKey = clientKey

	// this instead is a client unregistered for testing purposes
	s.UserClientUnregistered = &entities.Client{
		UserUUID:   s.UserFixture.UserUUID,
		ClientUUID: uuid.New().String(),
		Name:       "TEST",
		MachineID:  fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))),
	}

	// create raspberryPI instance first
	_, err = s.Service.Usecase.CreateRaspberryPI(s.UserClientRegistered.UserUUID, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))), "test")
	s.Require().NoError(err)

	// create handshake pending tasks
	handshakeID, err := s.Service.Usecase.CreateHandshake(s.UserClientRegistered.UserUUID, s.UserClientRegistered.Name, "XX:XX:XX:XX:XX:XX", constants.NothingStatus, utils.StringToBase64String("test.pcap"))
	s.Require().NoError(err)

	s.HandshakeValidID = handshakeID

	// assign handshake to client
	_, err = s.Service.Usecase.UpdateClientTask(s.UserClientRegistered.UserUUID, handshakeID, s.UserClientRegistered.ClientUUID, constants.PendingStatus, "", "", "")
	s.Require().NoError(err)

	temp, err := s.Client.Login(context.Background(), &pb.AuthRequest{
		Username: s.UserFixture.Username,
		Password: s.UserFixture.Password,
	})

	s.Require().NoError(err)
	s.UserTokenFixture = temp.Details

	temp, err = s.Client.Login(context.Background(), &pb.AuthRequest{
		Username: s.NormalUserFixture.Username,
		Password: s.NormalUserFixture.Password,
	})

	s.Require().NoError(err)
	s.NormalUserTokenFixture = temp.Details
}

// TearDownAllSuite implements suite.SetupTestSuite and is called after each suite
func (s *GRPCServerTestSuite) TearDownSuite() {
	// restore DB as its original state
	err := s.DatabaseUser.CleanDB([]string{entities.UserTableName})
	s.Require().NoError(err)

	err = s.DatabaseCert.CleanDB([]string{entities.CertTableName})
	s.Require().NoError(err)

	// this can be improved, creating a repository wrapper just for passing db instance is not the best
	// this needs RESET AND DEBUG ENV VARIABLES SET TO TRUE
	rr, err := repository.NewRepository(s.DatabaseUser, s.DatabaseCert)
	s.Require().NoError(err)

	err = seed.LoadUsers(rr)
	s.Require().NoError(err)

	err = s.DatabaseUser.CloseDatabase()
	s.Require().NoError(err)

	err = s.DatabaseCert.CloseDatabase()
	s.Require().NoError(err)
}
