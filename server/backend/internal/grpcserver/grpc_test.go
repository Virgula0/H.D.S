// nolint all
package grpcserver_test

import (
	"context"
	_ "context"
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
				Jwt:       s.TokenFixture,
				MachineId: "THISISATEST",
				Name:      "NEW CLIENT",
			},
			expectedOutput: &pb.GetClientInfoResponse{
				IsRegistered: false,
				Name:         "NEW CLIENT",
				MachineId:    "THISISATEST",
			},
		},
		{
			testname: "IsRegistered true on a client with MACHINE_ID already present in the table",
			request: &pb.GetClientInfoRequest{
				Jwt:       s.TokenFixture,
				Name:      s.UserClient.Name,
				MachineId: s.UserClient.MachineID,
			},
			expectedOutput: &pb.GetClientInfoResponse{
				IsRegistered: true,
				Name:         s.UserClient.Name,
				MachineId:    s.UserClient.MachineID,
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
