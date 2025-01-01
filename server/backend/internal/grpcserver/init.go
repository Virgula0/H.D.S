package grpcserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	GRPCCtx pb.HDSTemplateServiceServer
}

// New create new gRPC server-grpc with all required b. logic handlers
func New(ctx *ServerContext) *Server {
	return &Server{
		GRPCCtx: ctx,
	}
}

// Run start gRPC server-grpc listening on given port
func (s *Server) Run(ctx context.Context, opt *Option) error {
	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", opt.GrpcURL)
	if err != nil {
		return fmt.Errorf("gRPC server-grpc error, failed to listen: %v", err)
	}

	// grpc server-grpc options
	options := []grpc.ServerOption{
		grpc.ConnectionTimeout(opt.GrpcConnTimeout),
	}
	if opt.Debug {
		options = append(options, grpc.UnaryInterceptor(logInterceptor))
	}

	server := grpc.NewServer(options...)
	pb.RegisterHDSTemplateServiceServer(server, s.GRPCCtx)
	if opt.Debug {
		reflection.Register(server)
	}

	log.Infof("Running hds gRPC server %s", opt.GrpcURL)
	if err = server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC server-grpc: %v", err)
	}

	// clear connection
	err = lis.Close()
	if err != nil {
		return fmt.Errorf("failed to close listening of gRPC server-grpc: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	server.GracefulStop()
	return nil
}

// logInterceptor grpc debug purposes
func logInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, uHandler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	h, err := uHandler(ctx, req)
	log.Infof("Request - Method:%s\tDuration:%s\tError:%v\n", info.FullMethod, time.Since(start), err)
	return h, err
}
