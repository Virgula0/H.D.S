package raspberrypi_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"time"
)

func (s *ServerTCPIPSuite) Test_TCPServer_ConnectionButFailInCommand() {
	tests := []struct {
		testname       string
		expectedOutput string
	}{
		{
			testname:       "Valid name",
			expectedOutput: "FAIL\n",
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()
			errDead := client.SetDeadline(time.Now().Add(3 * time.Minute))
			s.Require().NoError(errDead, "Failed to set deadline")

			marshaled := []byte("hello")

			// Send length of data first
			ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err := client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			//ack
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")

			// Optional: Add a delay to simulate network conditions
			time.Sleep(10 * time.Millisecond)

			// Send the actual data
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send data")

			// read message
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, tt.expectedOutput, "unexpected response from server")
		})
	}
}

func (s *ServerTCPIPSuite) Test_TCPServer_LoginCommand() {
	tests := []struct {
		testname       string
		request        *entities.AuthRequest
		expectedOutput func(input string) bool
	}{
		{
			testname: "Valid login",
			request: &entities.AuthRequest{
				Username: s.UserFixture.Username,
				Password: s.UserFixture.Password,
			},
			expectedOutput: func(input string) bool {
				var pattern = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\n$"
				// Compile the regex pattern
				re, err := regexp.Compile(pattern)
				s.Require().NoError(err, "Failed to compile regexp")

				// Check if the input matches the pattern
				return re.MatchString(input)
			},
		},
		{
			testname: "Invalid username",
			request: &entities.AuthRequest{
				Username: utils.GenerateToken(32),
				Password: s.UserFixture.Password,
			},
			expectedOutput: func(input string) bool {
				var pattern = "invalid credentials"
				return strings.Contains(input, pattern)
			},
		},
		{
			testname: "Invalid password",
			request: &entities.AuthRequest{
				Username: s.UserFixture.Username,
				Password: utils.GenerateToken(32),
			},
			expectedOutput: func(input string) bool {
				var pattern = "invalid credentials"
				return strings.Contains(input, pattern)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()
			errDead := client.SetDeadline(time.Now().Add(3 * time.Minute))
			s.Require().NoError(errDead, "Failed to set deadline")

			marshaled := []byte("LOGIN")

			// Send length of data first
			ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err := client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			//ack
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")

			// Send the actual data
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send data")

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")

			marshaled, err = json.Marshal(tt.request)
			s.Require().NoError(err)

			// Send length of data first
			ll = []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err = client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")

			// Send the actual data
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send data")

			// ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")

			// check token
			response, err = bufio.NewReader(client).ReadString('\n')
			log.Println(response)
			s.Require().NoError(err, "Failed latest read from server")
			s.Require().True(tt.expectedOutput(response))
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
			testname: "JWT not valid",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           "test",
				MachineID:     utils.GenerateToken(32),
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				log.Println(response)
				return strings.Contains(response, "failed on the 'jwt' tag")
			},
		},
		{
			testname: "Error on machineID not valid",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
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
				MachineID:     utils.GenerateToken(32),
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "no valid handshakes provided")
			},
		},
		{
			testname: "Raspberrypi already present in the table for the correct user (IGNORE THE CREATION)",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.AdminToken,
				MachineID:     s.ExistingRaspberryMachineID,
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "no valid handshakes provided")
			},
		},
		{
			testname: "Raspberrypi already present also for another user (IGNORE THE CREATION)",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Jwt:           s.NormalUserToken,
				MachineID:     s.ExistingRaspberryMachineID,
				EncryptionKey: utils.GenerateToken(64),
			},
			expectedOutput: func(response string) bool {
				return strings.Contains(response, "no valid handshakes provided")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			client := s.Client()
			defer client.Close()
			errDead := client.SetDeadline(time.Now().Add(3 * time.Minute))
			s.Require().NoError(errDead, "Failed to set deadline")

			marshaled := []byte("HANDSHAKE")

			log.Println("sending size")
			// Send length of data first
			ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err := client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			//ack
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("sending content")
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send message")

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("sending size")

			marshaled, err = json.Marshal(tt.request)
			s.Require().NoError(err, "Failed to marshal request")

			// Send length of data first
			ll = []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err = client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("sending content")

			// Send the actual data
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send data")

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("reading content")
			// read response from server
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().True(tt.expectedOutput(response), "Condition not matched for "+tt.testname)
		})
	}
}

func (s *ServerTCPIPSuite) Test_TCPServer_TestOnHandshakeCreation() {
	var pcapTest = utils.StringToBase64String("test.pcap") // Optional: Add a delay to simulate network conditions
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
						ClientUUID:       nil,
						UUID:             "",
						SSID:             s.TestSSID,
						BSSID:            s.TestBSSID,
						UploadedDate:     "",
						Status:           constants.NothingStatus,
						CrackedDate:      nil,
						HashcatOptions:   nil,
						HashcatLogs:      nil,
						CrackedHandshake: nil,
						HandshakePCAP:    &pcapTest,
					},
				},
				Jwt:           s.AdminToken,
				MachineID:     s.ExistingRaspberryMachineID,
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
			testname: "Ok expecting 2 handshakes  to be created",
			request: &raspberrypi.TCPCreateRaspberryPIRequest{
				Handshakes: []*entities.Handshake{
					{
						ClientUUID:       nil,
						UUID:             "",
						SSID:             utils.GenerateToken(10),
						BSSID:            utils.GenerateToken(10),
						UploadedDate:     "",
						Status:           constants.NothingStatus,
						CrackedDate:      nil,
						HashcatOptions:   nil,
						HashcatLogs:      nil,
						CrackedHandshake: nil,
						HandshakePCAP:    &pcapTest,
					},
					{
						ClientUUID:       nil,
						UUID:             "",
						SSID:             utils.GenerateToken(10),
						BSSID:            utils.GenerateToken(10),
						UploadedDate:     "",
						Status:           constants.NothingStatus,
						CrackedDate:      nil,
						HashcatOptions:   nil,
						HashcatLogs:      nil,
						CrackedHandshake: nil,
						HandshakePCAP:    &pcapTest,
					},
				},
				Jwt:           s.AdminToken,
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
			/*
				Sleeps in this test emulate delay in connections
			*/
			client := s.Client()
			defer client.Close()
			errDead := client.SetDeadline(time.Now().Add(3 * time.Minute))
			s.Require().NoError(errDead, "Failed to set deadline")

			marshaled := []byte("HANDSHAKE")

			log.Println("sending size")
			// Send length of data first
			ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err := client.Write(ll)
			s.Require().NoError(err, "Failed to send size")
			time.Sleep(10 * time.Millisecond)

			//ack
			response, err := bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)
			time.Sleep(10 * time.Millisecond)

			log.Println("sending content")
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send message")
			time.Sleep(10 * time.Millisecond)

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("sending size")

			time.Sleep(10 * time.Millisecond)

			marshaled, err = json.Marshal(tt.request)
			s.Require().NoError(err, "Failed to marshal request")

			time.Sleep(10 * time.Millisecond)

			// Send length of data first
			ll = []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
			_, err = client.Write(ll)
			s.Require().NoError(err, "Failed to send size")

			time.Sleep(10 * time.Millisecond)

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			log.Println("sending content")

			time.Sleep(10 * time.Millisecond)

			// Send the actual data
			_, err = client.Write(marshaled)
			s.Require().NoError(err, "Failed to send data")

			time.Sleep(10 * time.Millisecond)

			//ack
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().Contains(response, "ACK\n", "unexpected response from server")
			log.Println(response)

			time.Sleep(10 * time.Millisecond)

			log.Println("reading content")
			// read response from server
			response, err = bufio.NewReader(client).ReadString('\n')
			s.Require().NoError(err, "Failed to read from server")
			s.Require().True(tt.expectedOutput(response), "Condition not matched for "+tt.testname)
		})
	}
}
