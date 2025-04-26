package grpcserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
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

	serverCert, err := tls.X509KeyPair(opt.ServerCert, opt.ServerKey)
	if err != nil {
		return fmt.Errorf("failed to load server certificate and key: %v", err)
	}

	// Create a certificate pool and add the CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(opt.CACert) {
		return errors.New("failed to append CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{serverCert},
		ClientCAs:          certPool, // verify client cert from server
		ClientAuth:         tls.RequireAnyClientCert,
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: false,
		GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			// Get the client configuration
			log.Infof("[gRPC] Client Name (ID) %s", info.ServerName)
			clientConfig, exists := opt.ClientConfigStorage.GetClientConfig(info.ServerName)

			nocert := &tls.Config{
				Certificates: []tls.Certificate{serverCert},
				ClientCAs:    certPool,
				MinVersion:   tls.VersionTLS13,
				ClientAuth:   tls.NoClientCert, // Allows clients to connect without certificates
			}

			if !exists {
				log.Warnf("No config found for client: %s", info.ServerName)
				return nocert, nil // Allow plaintext by default if no config is found
			}

			if !exists && clientConfig.EncryptionEnabled {
				log.Errorf("No config found for client: %s but the encryption is enabled, delete certs from client directory and retry", info.ServerName)
				return nocert, fmt.Errorf("no config found for client: %s but the encryption is enabled, delete certs from client directory and retry", info.ServerName) // Allow plaintext by default if no config is found
			}

			if !clientConfig.EncryptionEnabled {
				log.Warnf("Encryption disabled for client: %s; allowing plaintext connection", clientConfig.ID)
				// Return a nil config to allow plaintext
				return nocert, nil
			}

			log.Infof("[gRPC] Encryption enabled for %s", info.ServerName)

			// If encryption is enabled, create a new TLS config that requires mutual authentication
			return &tls.Config{
				Certificates: []tls.Certificate{serverCert},
				ClientCAs:    certPool,
				ClientAuth:   tls.RequireAndVerifyClientCert,
				MinVersion:   tls.VersionTLS13,
			}, nil
		},
	}

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", opt.GrpcURL)
	if err != nil {
		return fmt.Errorf("gRPC server-grpc error, failed to listen: %v", err)
	}

	// grpc server-grpc options
	options := []grpc.ServerOption{
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.ConnectionTimeout(opt.GrpcConnTimeout),
		grpc.MaxRecvMsgSize(customErrors.MaxUploadSize), // limit for receiving
		grpc.MaxSendMsgSize(customErrors.MaxUploadSize), // limit for sending
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
