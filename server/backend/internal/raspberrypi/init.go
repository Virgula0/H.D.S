package raspberrypi

import (
	"fmt"
	handlers "github.com/Virgula0/progetto-dp/server/backend/internal/restapi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"net"
	"time"
)

type TCPServer struct {
	l         net.Listener
	timeout   time.Duration
	sleepTime time.Duration
	usecase   *usecase.Usecase
	TCPHandler
}

// NewTCPServer creates and returns a new TCPServer instance, initializing it with the provided TCP connection and usecase.
func NewTCPServer(service *handlers.ServiceHandler, address, port string) (*TCPServer, error) {
	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))

	if err != nil {
		return nil, err
	}

	return &TCPServer{
		l:         conn,
		usecase:   service.Usecase,
		timeout:   30 * time.Second,
		sleepTime: 200 * time.Millisecond,
	}, nil
}
