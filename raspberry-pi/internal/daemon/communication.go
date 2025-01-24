package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/enums"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

/*
Authenticator

Uses provided credentials for authenticating the user via REST/API each hour
*/
func (r *RaspberryPiInfo) Authenticator(wr *Communicate) {
	ticker := time.NewTicker(1 * time.Hour) // every hour

	for {
		// 1. Send login request
		err := wr.writeToServerCommand(enums.LOGIN)

		if err != nil {
			log.Fatalf("[RSP-PI] Failed to write command to the server: %s", err.Error())
		}

		if jwt, err := wr.writeToServerAuthRequest(r.Credentials); err == nil {
			*r.JWT = jwt
			r.FirstLogin <- true
		} else {
			log.Fatal(err)
		}

		<-ticker.C
	}
}

type Communicate struct {
	*Client
}

func (c *Client) readFromServer() (string, error) {
	// bufio.NewReader(c.Conn).ReadString('\n') this includes the \n at the end as result
	ss, err := bufio.NewReader(c.Conn).ReadString('\n')

	if err != nil {
		return "", err
	}

	return strings.TrimRight(ss, "\n"), nil
}

func (c *Client) isACKMessage(msg string) bool {
	return msg == enums.ACK.String()
}

func (c *Client) writeToServerCommand(command enums.Command) error {
	cc := command.String()
	ll := []byte(fmt.Sprintf("%v", len(cc)) + "\n")
	_, err := c.Conn.Write(ll)

	if err != nil {
		return err
	}

	// accept ack from the server
	err = c.readACKFromServer()
	if err != nil {
		// if here command not valid
		return err
	}

	// Send the actual data
	_, err = c.Conn.Write([]byte(cc + "\n"))

	if err != nil {
		return err
	}

	// accept ack from the server
	err = c.readACKFromServer()
	if err != nil {
		// if here command not valid
		return err
	}

	return nil
}

func (c *Client) writeToServerAuthRequest(request *entities.AuthRequest) (string, error) {
	marshaled, err := json.Marshal(request)

	if err != nil {
		return "", err
	}

	ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
	_, err = c.Conn.Write(ll)

	if err != nil {
		return "", err
	}

	// accept ack from the server
	err = c.readACKFromServer()

	if err != nil {
		return "", err
	}

	// Send the actual data
	_, err = c.Conn.Write(marshaled)

	if err != nil {
		return "", err
	}

	// accept ack from the server
	err = c.readACKFromServer()

	if err != nil {
		return "", err
	}

	// read token
	token, err := c.readFromServer()
	if err != nil {
		return "", err

	}

	return token, nil
}

func (c *Client) writeToServerHandshake(request entities.TCPCreateRaspberryPIRequest) (int, error) {
	marshaled, err := json.Marshal(request)

	if err != nil {
		return -1, err
	}

	ll := []byte(fmt.Sprintf("%v", len(marshaled)) + "\n")
	_, err = c.Conn.Write(ll)

	if err != nil {
		return 0, err
	}

	// accept ack from the server
	err = c.readACKFromServer()

	if err != nil {
		return 0, err
	}

	// Send the actual data
	wrote, err := c.Conn.Write(marshaled)

	if err != nil {
		return 0, err
	}

	// accept ack from the server
	err = c.readACKFromServer()

	if err != nil {
		return 0, err
	}

	return wrote, nil
}

func (c *Client) readACKFromServer() error {
	// accept ack from the server
	read, err := c.readFromServer()

	if err != nil {
		return err
	}

	if !c.isACKMessage(read) {
		var status = enums.FAIL
		return fmt.Errorf("error did not received an ACK %s", status)
	}

	return nil
}

// HandleServerCommunication handles data exchange with the server.
func (c *Communicate) HandleServerCommunication(instance *RaspberryPiInfo, machineID string, handshakes []*entities.Handshake) error {
	// 1. Send hadnshake request
	err := c.writeToServerCommand(enums.HANDSHAKE)

	if err != nil {
		return fmt.Errorf("[RSP-PI] Failed to write command to the server: %s", err.Error())
	}

	request := entities.TCPCreateRaspberryPIRequest{
		Handshakes:    handshakes,
		Jwt:           *instance.JWT,
		MachineID:     machineID,
		EncryptionKey: "",
	}

	wrote, err := c.writeToServerHandshake(request)
	if err != nil {
		return fmt.Errorf("[RSP-PI] Failed to write to server: %s", err.Error())
	}
	log.Printf("[RSP-PI] Wrote %v bytes", wrote)

	response, err := c.readFromServer()
	if err != nil {
		return fmt.Errorf("[RSP-PI] Failed to write to server: %s", err.Error())
	}

	log.Println("[RSP-PI] Response from server:", response)

	return nil
}
