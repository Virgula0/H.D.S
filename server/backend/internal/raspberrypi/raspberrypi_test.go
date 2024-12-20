package raspberrypi_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
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
				return strings.Contains(response, "No handshakes provided")
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
				return strings.Contains(response, "No handshakes provided")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()

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

func sendChunkedData(client net.Conn, data []byte, chunkSize int) error {
	totalBytes := len(data)
	for i := 0; i < totalBytes; i += chunkSize {
		end := i + chunkSize
		if end > totalBytes {
			end = totalBytes
		}

		chunk := data[i:end]

		// Send chunk size first (optional but can be useful for the server to handle the data properly)
		_, err := client.Write([]byte(fmt.Sprintf("%d\n", len(chunk)))) // Sending chunk length as a header
		if err != nil {
			return fmt.Errorf("failed to send chunk size: %w", err)
		}

		// Now send the actual chunk of data
		_, err = client.Write(chunk)
		if err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}

		// Optionally wait for server acknowledgment if necessary
		// You can implement acknowledgment logic here if your server expects it
	}

	// Optionally send an end-of-data marker
	_, err := client.Write([]byte("EOF\n"))
	return err
}
func (s *ServerTCPIPSuite) Test_TCPServer_TestOnHandshakeCreation() {
	var pcapTest = utils.StringToBase64String("ciao")
	var a string
	var b string
	var c string
	var d string
	var e string

	tests := []struct {
		testname       string
		request        *raspberrypi.TCPCreateRaspberryPIRequest
		expectedOutput func(response string) bool
	}{
		{
			testname: "Ok expecting handshake to be created",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Handshakes: []*entities.Handshake{
					{
						UserUUID:         s.UserFixture.UserUUID,
						ClientUUID:       &a,
						RaspberryPIUUID:  s.RaspberryPIExistingID,
						UUID:             "",
						SSID:             s.TestSSID,
						BSSID:            s.TestBSSID,
						UploadedDate:     "",
						Status:           constants.NothingStatus,
						CrackedDate:      &b,
						HashcatOptions:   &c,
						HashcatLogs:      &d,
						CrackedHandshake: &e,
						HandshakePCAP:    &pcapTest,
					},
				},
				Jwt:           s.AdminToken,
				UserUUID:      s.UserFixture.UserUUID,
				MachineID:     s.ExistingRaspberryMachineID,
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				regex := "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12};"
				matched, err := regexp.Match(regex, []byte(response))
				s.Require().NoError(err, "Id seems to be not an uuid")
				return matched
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()

			marshaled, err := json.Marshal(tt.request)
			s.Require().NoError(err, "Failed to marshal request")

			chunkSize := 1024 // Example: send 1KB chunks

			err = sendChunkedData(client, marshaled, chunkSize)
			s.Require().NoError(err, "Failed to send chunked data")

			// read response from server
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")

			log.Println("Response from server: " + response)

			s.Require().True(tt.expectedOutput(response), "Condition not matched for "+tt.testname)
		})
	}
}
