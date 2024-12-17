package grpc_server_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

func Test_GRPC_Connection(t *testing.T) {
	/*
		GRPC Server must run!
	*/
	serverAddress := "localhost:7777"

	// Connect to the gRPC server
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to gRPC server")
	defer conn.Close()

	client := pb.NewHDSTemplateServiceClient(conn)

	// Define test cases
	tests := []struct {
		name           string
		request        *pb.HelloRequest
		expectedOutput *pb.HelloResponse
	}{
		{
			name:           "Valid name",
			request:        &pb.HelloRequest{Name: "World"},
			expectedOutput: &pb.HelloResponse{Message: "Hello, World!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
