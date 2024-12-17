// nolint all
package grpcserver_test

import (
	"context"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/testsuite"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

const serverAddress = "localhost:7777"

func newClientConn(t *testing.T) (*grpc.ClientConn, pb.HDSTemplateServiceClient) {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to gRPC server")

	return conn, pb.NewHDSTemplateServiceClient(conn)
}

func Test_GRPC_Connection(t *testing.T) {
	/*
		GRPC Server must run!
	*/
	// Connect to the gRPC server
	conn, client := newClientConn(t)
	defer conn.Close()

	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.HelloRequest
		expectedOutput *pb.HelloResponse
	}{
		{
			testname:       "Valid name",
			request:        &pb.HelloRequest{Name: "World"},
			expectedOutput: &pb.HelloResponse{Message: "Hello, World!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			// Perform the gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.Test(ctx, tt.request)

			// Use require to assert no errors in the RPC call
			require.NoError(t, err, "Test RPC failed")

			// Use assert to check the response
			assert.Equal(t, tt.expectedOutput.Message, resp.Message, "Unexpected response from Test RPC")
		})
	}
}

func Test_GetClientInfo_Method(t *testing.T) {
	/*
		GRPC Server must run!
	*/
	// Connect to the gRPC server
	conn, client := newClientConn(t)
	defer conn.Close()

	// Perform login and get valid token first
	token, err := testsuite.AuthAPI(authenticate.AuthRequest{
		Username: seed.UserSeed.AdminUser.Username,
		Password: seed.UserSeed.AdminUser.Password,
	})
	require.NoError(t, err)

	// Define test cases
	tests := []struct {
		testname       string
		request        *pb.GetClientInfoRequest
		expectedOutput *pb.GetClientInfoResponse
	}{
		{
			testname: "Non Registered Client OK",
			request: &pb.GetClientInfoRequest{
				Jwt:       token,
				MachineId: "THISISATEST",
				Name:      "NEW CLIENT",
			},
			expectedOutput: &pb.GetClientInfoResponse{
				IsRegistered: false,
				Name:         "NEW CLIENT",
				MachineId:    "THISISATEST",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			// Perform the gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.GetClientInfo(ctx, tt.request)

			// Use require to assert no errors in the RPC call
			require.NoError(t, err, "Test RPC failed")

			// Use assert to check the response
			assert.Equal(t, tt.expectedOutput.IsRegistered, resp.IsRegistered, "Unexpected response from Test RPC")
			assert.Equal(t, tt.expectedOutput.Name, resp.GetName(), "Unexpected response from Test RPC")
			assert.Equal(t, tt.expectedOutput.MachineId, resp.GetMachineId(), "Unexpected response from Test RPC")
		})
	}
}
