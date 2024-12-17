package grpc_server

import (
	"context"
	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

// Test test is just useful for checking correct gRPC server stage
func (s *ServerContext) Test(_ context.Context, pbRequest *pb.HelloRequest) (*pb.HelloResponse, error) {
	// do usecase stuff
	return &pb.HelloResponse{
		Message: "Hello, World!",
	}, nil
}
