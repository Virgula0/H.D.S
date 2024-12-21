package grpcserver

import (
	usecaseHandler "github.com/Virgula0/progetto-dp/server/backend/internal/usecase"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

type ServerContext struct {
	Usecase *usecaseHandler.Usecase
	pb.UnimplementedHDSTemplateServiceServer
}

// NewServerContext Inject context into GRPC SERVER
func NewServerContext(usecase *usecaseHandler.Usecase) *ServerContext {
	return &ServerContext{
		Usecase: usecase,
	}
}
