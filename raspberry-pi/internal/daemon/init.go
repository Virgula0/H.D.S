package daemon

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/constants"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	"net"
	"time"
)

var serverTimedOutDuration = time.Second * 30

type RaspberryPiInfo struct {
	JWT         *string
	FirstLogin  chan bool
	Credentials *entities.AuthRequest
}

type Client struct {
	Conn net.Conn
}

func InitClientConnection() (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", constants.TCPAddress, constants.TCPPort))
	if err != nil {
		return nil, err
	}

	err = conn.SetDeadline(time.Now().Add(serverTimedOutDuration))
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn: conn,
	}, nil
}
