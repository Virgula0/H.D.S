package raspberrypi_test

import (
	"bufio"
	"encoding/json"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

func (s *ServerTCPIPSuite) Test_TCPServer_Connection() {

	client := s.Client()
	defer client.Close()

	tests := []struct {
		testname       string
		expectedOutput string
	}{
		{
			testname:       "Valid name",
			expectedOutput: "invalid character 'h' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			// Perform the gRPC request

			_, err := client.Write([]byte("hello"))
			s.Require().NoError(err, "Failed to write to server")

			// read response from server
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")

			s.Require().Contains(response, tt.expectedOutput, "Cannot unmarshal TCPCreateRaspberryPIRequest from client")
		})
	}
}

func (s *ServerTCPIPSuite) Test_TCPServer_CreateRaspberryPIValidation() {

	tests := []struct {
		testname       string
		request        *raspberrypi.TCPCreateRaspberryPIRequest
		expectedOutput func(response string) bool
	}{
		{
			testname: "User UUID not valid",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				UserUUID:      "test",
				MachineID:     utils.GenerateToken(32),
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "Error:Field validation for 'UserUUID'")
			},
		},
		{
			testname: "Error on machineID not valid",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				UserUUID:      s.UserFixture.UserUUID,
				MachineID:     utils.GenerateToken(64),
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "Error:Field validation for 'MachineID'")
			},
		},
		{
			testname: "Error on machineID not valid",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				UserUUID:      s.UserFixture.UserUUID,
				MachineID:     utils.GenerateToken(32),
				EncryptionKey: utils.GenerateToken(32),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "Error:Field validation for 'EncryptionKey'")
			},
		},
		{
			testname: "RaspberryPI created",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				UserUUID:      uuid.New().String(),
				MachineID:     utils.GenerateToken(32),
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				regex := "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
				matched, err := regexp.Match(regex, []byte(response))
				s.Require().NoError(err, "Id seems to be not an uuid")
				return matched
			},
		},
		{
			testname: "Raspberrypi already present in the table for the correct user (IGNORE THE CREATION)",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				UserUUID:      uuid.New().String(),
				MachineID:     s.ExistingRaspberryMachineID,
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "Duplicate entry")
			},
		},
		{
			testname: "Raspberrypi already present also for another user (IGNORE THE CREATION)",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.NormalUserToken,
				UserUUID:      uuid.New().String(),
				MachineID:     s.ExistingRaspberryMachineID,
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "Duplicate entry")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()

			// Perform the gRPC request

			marshaled, err := json.Marshal(tt.request)
			s.Require().NoError(err, "Failed to marshal request")

			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to write to server")

			// read response from server
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")

			log.Println("Response from server " + response)

			s.Require().True(tt.expectedOutput(response), "Condition not matched for "+tt.testname)
		})
	}
}
