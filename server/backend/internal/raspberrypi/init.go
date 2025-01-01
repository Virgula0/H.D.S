package raspberrypi

import (
	"fmt"
	handlers "github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"net"
	"time"
)

type TCPServer struct {
	w       net.Listener
	timeout time.Duration
	usecase *usecase.Usecase
	TCPHandler
}

// NewTCPServer creates and returns a new TCPServer instance, initializing it with the provided TCP connection and usecase.
func NewTCPServer(service *handlers.ServiceHandler, address, port string) (*TCPServer, error) {
	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))

	if err != nil {
		return nil, err
	}

	return &TCPServer{
		w:       conn,
		usecase: service.Usecase,
		timeout: 30 * time.Second,
	}, nil
}
