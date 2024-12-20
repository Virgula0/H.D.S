// nolint all
package raspberrypi_test

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/testsuite"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServerTCPIPSuite struct {
	testsuite.TCPServerSuite
	UserFixture          *entities.User
	UserClientRegistered *entities.Client

	ExistingRaspberryMachineID string
	AdminToken                 string
	NormalUser                 *entities.User
	NormalUserToken            string
	RaspberryPIExistingID      string
	TestSSID                   string
	TestBSSID                  string
}

// Run All tests
func TestTCPMethodCaller(t *testing.T) {
	suite.Run(t, new(ServerTCPIPSuite))
}

func (s *ServerTCPIPSuite) SetupSuite() {
	s.TCPServerSuite.SetupSuite()

	s.Require().NotNil(seed.AdminSeed.User)
	s.UserFixture = seed.AdminSeed.User

	s.Require().NotNil(seed.NormalUserSeed.User)
	s.NormalUser = seed.NormalUserSeed.User

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

	// create raspberryPI instance first
	machineID := fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10))))
	respID, err := s.Service.Usecase.CreateRaspberryPI(s.UserClientRegistered.UserUUID, machineID, "test")
	s.Require().NoError(err)

	s.ExistingRaspberryMachineID = machineID
	s.RaspberryPIExistingID = respID

	s.AdminToken, err = testsuite.AuthAPI(authenticate.AuthRequest{
		Username: s.UserFixture.Username,
		Password: s.UserFixture.Password,
	})
	s.Require().NoError(err)

	s.NormalUserToken, err = testsuite.AuthAPI(authenticate.AuthRequest{
		Username: s.NormalUser.Username,
		Password: s.NormalUser.Password,
	})
	s.Require().NoError(err)

	s.TestSSID = "TEST"
	s.TestSSID = "XX:XX:XX:XX:XX:XX"

}

// TearDownAllSuite implements suite.SetupTestSuite and is called after each suite
func (s *ServerTCPIPSuite) TearDownSuite() {
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
