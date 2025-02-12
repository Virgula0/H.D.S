// nolint all
package grpcserver_test

import (
	"context"
	_ "context"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
	_ "time"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

func (s *GRPCServerTestSuite) Test_GRPC_Connection() {
	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.HelloRequest
		expectedOutput *pb.HelloResponse
	}{
		{
			testname:       "Valid name",
			request:        &pb.HelloRequest{Name: "Hello, World"},
			expectedOutput: &pb.HelloResponse{Message: "Hello, World!"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			// Perform the gRPC request
			client := s.Client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := client.Test(ctx, tt.request)

			// Use require to assert no errors in the RPC call
			s.Require().NoError(err, "Test RPC failed")

			// Use assert to check the response

			s.Require().Equal(tt.expectedOutput.Message, resp.Message, "Unexpected response from Test RPC")
		})
	}
}

func (s *GRPCServerTestSuite) Test_GRPC_Login() {
	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.AuthRequest
		expectedOutput func(resp *pb.UniformResponse, err error) bool
	}{
		{
			testname: "Valid username",
			request: &pb.AuthRequest{
				Username: s.UserFixture.Username,
				Password: s.UserFixture.Password,
			},
			expectedOutput: func(resp *pb.UniformResponse, err error) bool {
				s.Require().NoError(err, "Test RPC failed")
				return resp.Details != "" && resp.Status == "logged_in"
			},
		},
		{
			testname: "Not Valid Username",
			request: &pb.AuthRequest{
				Username: "test",
				Password: s.UserFixture.Password,
			},
			expectedOutput: func(resp *pb.UniformResponse, err error) bool {
				s.Require().Contains(err.Error(), "invalid credentials", "invalid credentials fail")
				return true
			},
		},
		{
			testname: "Not Valid Password",
			request: &pb.AuthRequest{
				Username: s.UserFixture.Username,
				Password: utils.GenerateToken(10),
			},
			expectedOutput: func(resp *pb.UniformResponse, err error) bool {
				s.Require().Contains(err.Error(), "invalid credentials", "invalid credentials fail")
				return true
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			// Perform the gRPC request
			client := s.Client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := client.Login(ctx, tt.request)
			log.Println(err)
			// Use assert to check the response
			s.Require().True(tt.expectedOutput(resp, err))
		})
	}
}

func (s *GRPCServerTestSuite) Test_GetClientInfo_Method() {

	// Connect to the gRPC server
	client := s.Client

	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.GetClientInfoRequest
		expectedOutput *pb.GetClientInfoResponse
	}{
		{
			testname: "Non Registered Client",
			request: &pb.GetClientInfoRequest{
				Jwt:       s.UserTokenFixture,
				MachineId: s.UserClientUnregistered.MachineID,
				Name:      s.UserClientUnregistered.Name,
			},
			expectedOutput: &pb.GetClientInfoResponse{
				IsRegistered: false,
				MachineId:    s.UserClientUnregistered.MachineID,
				Name:         s.UserClientUnregistered.Name,
			},
		},
		{
			testname: "IsRegistered true on a client with MACHINE_ID already present in the table",
			request: &pb.GetClientInfoRequest{
				Jwt:       s.UserTokenFixture,
				Name:      s.UserClientRegistered.Name,
				MachineId: s.UserClientRegistered.MachineID,
			},
			expectedOutput: &pb.GetClientInfoResponse{
				IsRegistered: true,
				Name:         s.UserClientRegistered.Name,
				MachineId:    s.UserClientRegistered.MachineID,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			// Perform the gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.GetClientInfo(ctx, tt.request)

			// Use require to assert no errors in the RPC call
			s.Require().NoError(err, "Test RPC failed")

			// Use assert to check the response
			s.Require().Equal(tt.expectedOutput.IsRegistered, resp.IsRegistered, "Unexpected response from Test RPC")
			s.Require().Equal(tt.expectedOutput.Name, resp.GetName(), "Unexpected response from Test RPC")
			s.Require().Equal(tt.expectedOutput.MachineId, resp.GetMachineId(), "Unexpected response from Test RPC")
		})
	}
}

func (s *GRPCServerTestSuite) Test_HashcatMessageService_Method() {

	// Connect to the gRPC server
	client := s.Client

	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.ClientTaskMessageFromClient
		expectedOutput *pb.ClientTaskMessageFromServer
	}{
		{
			testname: "Expect input task from server",
			request: &pb.ClientTaskMessageFromClient{
				Jwt: s.UserTokenFixture,
			},
			expectedOutput: &pb.ClientTaskMessageFromServer{
				Tasks: []*pb.ClientTask{
					{
						UserId:     s.UserClientRegistered.UserUUID,
						ClientUuid: s.UserClientRegistered.ClientUUID,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testname, func() {
			// Perform the gRPC request

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // timeout after 10 seconds
			defer cancel()

			stream, err := client.HashcatTaskChat(ctx)
			s.Require().NoError(err, "Stream initialization failed")

			var clientID string
			var crackedDate string
			var hashcatOptions string
			var hashcatLogs string
			var crackedHandshake string
			var handshakePCAP string

			var handshake = entities.Handshake{
				UserUUID:         "",
				ClientUUID:       &clientID,
				UUID:             "",
				SSID:             "",
				BSSID:            "",
				UploadedDate:     "",
				Status:           "",
				CrackedDate:      &crackedDate,
				HashcatOptions:   &hashcatOptions,
				HashcatLogs:      &hashcatLogs,
				CrackedHandshake: &crackedHandshake,
				HandshakePCAP:    &handshakePCAP,
			}

			// Start receiving messages from the server
			for {
				msg, err := stream.Recv()
				s.Require().NoError(err, "Failed to receive message from stream")

				// if this check is ok it means that the server is asking for starting the task
				// the client uuid is a known information since we will perform a GetClientInfo request
				// before
				for _, task := range msg.GetTasks() {
					if task.GetClientUuid() == s.UserClientRegistered.ClientUUID && task.GetStartCracking() {
						*handshake.HandshakePCAP = task.GetHashcatPcap()
						*handshake.ClientUUID = task.GetClientUuid()
						handshake.UUID = task.GetHandshakeUuid()
						handshake.UserUUID = task.GetUserId()
					}
				}
				if handshake.ClientUUID != nil {
					break
				}
			}

			// Use assert to check the response
			s.Require().Equal(handshake.UserUUID, tt.expectedOutput.Tasks[0].UserId, "Unexpected response from Test RPC")
			s.Require().Equal(*handshake.ClientUUID, tt.expectedOutput.Tasks[0].ClientUuid, "Unexpected response from Test RPC")
		})
	}
}

func (s *GRPCServerTestSuite) Test_HashcatMessageService_TimeoutOnNonRegisteredClient() {
	// Connect to the gRPC server
	client := s.Client

	// Define test case
	testName := "Ensure no task is sent for unregistered client"

	s.Run(testName, func() {
		// Set a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3) // Timeout after 3 seconds
		defer cancel()

		// Initialize the gRPC stream
		stream, err := client.HashcatTaskChat(ctx)
		s.Require().NoError(err, "Stream initialization failed")

		// Attempt to receive responses
		for {
			msg, err := stream.Recv()
			if err != nil {
				// Check if the error is a timeout or end-of-stream
				s.Require().Equal(context.DeadlineExceeded, ctx.Err(), "Expected timeout while waiting for server response")
				break
			}

			// Inspect the tasks sent by the server
			for _, task := range msg.GetTasks() {
				if task.GetClientUuid() == s.UserClientUnregistered.ClientUUID {
					s.FailNow("Server sent a task for an unregistered client", "Task: %v", task)
				}
			}
		}
	})
}

func (s *GRPCServerTestSuite) Test_HashcatMessageService_ErrorWhenClientTriesToUpdateAHashcatRowOfAnotherUser() {
	// Connect to the gRPC server
	client := s.Client

	testName := "Error when client tries to update a hashcat row of another user"
	request := &pb.ClientTaskMessageFromClient{
		Jwt:        s.NormalUserTokenFixture,
		ClientUuid: s.UserClientRegistered.ClientUUID, // Registered and valid client HandshakeUUID!
	}

	s.Run(testName, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20) // Timeout after 20 seconds
		defer cancel()

		stream, err := client.HashcatTaskChat(ctx)
		s.Require().NoError(err, "Stream initialization failed")

		err = stream.Send(request)
		s.Require().NoError(err, "Failed to send request to the server")

		response, recvErr := stream.Recv()

		if recvErr != nil {
			// Ensure the server returns the correct error when a user tries to update another user's row
			s.Require().Equal(codes.Internal, status.Code(recvErr), "Unexpected error code")
			s.Require().Contains(recvErr.Error(), "Cannot update client task ->  not found", "Unexpected error message")
		} else {
			s.FailNow("Unexpected response received: %v", response)
		}
	})
}

func (s *GRPCServerTestSuite) Test_HashcatMessageService_UpdateClientTaskSuccessfully() {
	// Connect to the gRPC server
	client := s.Client

	testName := "Client should be able to update its own info about its handshakes"
	request := &pb.ClientTaskMessageFromClient{
		Jwt:            s.UserTokenFixture,
		HandshakeUuid:  s.HandshakeValidID,
		ClientUuid:     s.UserClientRegistered.ClientUUID,
		HashcatOptions: "updated",
	}

	s.Run(testName, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20) // Timeout after 20 seconds
		defer cancel()

		stream, err := client.HashcatTaskChat(ctx)
		s.Require().NoError(err, "Stream initialization failed")

		err = stream.Send(request)
		s.Require().NoError(err, "Failed to send request to the server")
	})
}
