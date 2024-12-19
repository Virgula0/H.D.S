// notlint: all
package grpcserver_test

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/testsuite"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type GRPCServerTestSuite struct {
	testsuite.TestSuite
	UserFixture            *entities.User
	UserClientRegistered   *entities.Client
	UserClientUnregistered *entities.Client
	TokenFixture           string

	NormalUserFixture      *entities.User
	NormalUserTokenFixture string
	HandshakeValidID       string
}

// Run All tests
func TestGRPCMethodCaller(t *testing.T) {
	suite.Run(t, new(GRPCServerTestSuite))
}

func (s *GRPCServerTestSuite) SetupSuite() {
	s.TestSuite.SetupSuite()
	s.Require().NotNil(seed.AdminSeed.User)
	s.UserFixture = seed.AdminSeed.User

	s.Require().NotNil(seed.NormalUserSeed.User)
	s.NormalUserFixture = seed.NormalUserSeed.User

	// Create a client known
	s.UserClientRegistered = &entities.Client{
		UserUUID:  s.UserFixture.UserUUID,
		Name:      "TEST",
		MachineID: fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))),
	}

	clientID, err := s.Service.Usecase.CreateClient(s.UserFixture.UserUUID, s.UserClientRegistered.MachineID, s.UserClientRegistered.LatestIP, s.UserClientRegistered.Name)
	s.Require().NoError(err)

	// assign generated clientID
	s.UserClientRegistered.ClientUUID = clientID

	// this instead is a client unregistered for testing purposes
	s.UserClientUnregistered = &entities.Client{
		UserUUID:   s.UserFixture.UserUUID,
		ClientUUID: uuid.New().String(),
		Name:       "TEST",
		MachineID:  fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))),
	}

	// create raspberryPI instance first
	raspID, err := s.Service.Usecase.CreateRaspberryPI(s.UserClientRegistered.UserUUID, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))), "test")
	s.Require().NoError(err)

	// create handshake pending tasks
	handshakeID, err := s.Service.Usecase.CreateHandshake(s.UserClientRegistered.UserUUID, raspID, s.UserClientRegistered.Name, "XX:XX:XX:XX:XX:XX", constants.NothingStatus, utils.StringToBase64String("test.pcap"))
	s.Require().NoError(err)

	s.HandshakeValidID = handshakeID

	// assign handshake to client
	_, err = s.Service.Usecase.UpdateClientTask(s.UserClientRegistered.UserUUID, handshakeID, s.UserClientRegistered.ClientUUID, constants.PendingStatus, "", "", "")
	s.Require().NoError(err)

	s.TokenFixture, err = testsuite.AuthAPI(authenticate.AuthRequest{
		Username: s.UserFixture.Username,
		Password: s.UserFixture.Password,
	})
	s.Require().NoError(err)

	s.NormalUserTokenFixture, err = testsuite.AuthAPI(authenticate.AuthRequest{
		Username: s.NormalUserFixture.Username,
		Password: s.NormalUserFixture.Password,
	})
	s.Require().NoError(err)
}

// TearDownAllSuite implements suite.SetupTestSuite and is called after each suite
func (s *GRPCServerTestSuite) TearDownSuite() {
	// restore DB as its original state
	err := s.Database.CleanDB([]string{entities.UserTableName})
	s.Require().NoError(err)

	// this can be improved, creating a repository wrapper just for passing db instance is not the best
	rr, err := repository.NewRepository(s.Database)
	s.Require().NoError(err)

	err = seed.LoadUsers(rr)
	s.Require().NoError(err)

	err = s.Database.CloseDatabase()
	s.Require().NoError(err)
}
