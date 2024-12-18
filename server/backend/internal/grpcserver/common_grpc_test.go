package grpcserver_test

import (
	"crypto/md5"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/testsuite"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type GRPCServerTestSuite struct {
	testsuite.TestSuite
	UserFixture  *entities.User
	UserClient   *entities.Client
	TokenFixture string
}

// Run All tests
func TestGRPCMethodCaller(t *testing.T) {
	suite.Run(t, new(GRPCServerTestSuite))
}

func (s *GRPCServerTestSuite) SetupSuite() {
	s.TestSuite.SetupSuite()
	s.Require().NotNil(seed.AdminSeed.User)
	s.UserFixture = seed.AdminSeed.User

	// Create a client known
	s.UserClient = &entities.Client{
		UserUUID:   s.UserFixture.UserUUID,
		ClientUUID: uuid.New().String(),
		Name:       "TEST",
		MachineID:  fmt.Sprintf("%x", md5.Sum([]byte(s.UserFixture.UserUUID))),
	}
	_, err := s.Service.Usecase.CreateClient(s.UserFixture.UserUUID, s.UserClient.MachineID, s.UserClient.LatestIP, s.UserClient.Name)
	s.Require().NoError(err)

	s.TokenFixture, err = testsuite.AuthAPI(authenticate.AuthRequest{
		Username: s.UserFixture.Username,
		Password: s.UserFixture.Password,
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
