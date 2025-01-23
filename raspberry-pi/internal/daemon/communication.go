package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/authapi"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	log "github.com/sirupsen/logrus"
	"time"
)

type ServerStatus int

const (
	ACK = iota + 1
	FAIL
)

func (s ServerStatus) String() string {
	return [...]string{"ACK\n", "FAIL\n"}[s-1]
}

func (s ServerStatus) EnumIndex() int {
	return int(s)
}

/*
Authenticator

Uses provided credentials for authenticating the user via REST/API each hour
*/
func (r *RaspberryPiInfo) Authenticator() {
	ticker := time.NewTicker(1 * time.Hour) // every hour

	for {
		// Every hour re-authenticate
		if jwt, err := authapi.AuthAPI(r.Credentials); err == nil {
			*r.JWT = jwt
			r.FirstLogin <- true
		} else {
			log.Fatal(err)
		}

		<-ticker.C
	}
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

func (c *Client) isACKMessage(msg string) bool {
	var ack ServerStatus = ACK
	return msg == ack.String()
}

func (c *Client) readACKFromServer() error {
	// accept ack from the server
	read, err := c.readFromServer()

	if err != nil {
		return err
	}

	if !c.isACKMessage(read) {
		var status ServerStatus = FAIL
		return fmt.Errorf("error ACK from the server %s", status)
	}

	return nil
}

func (c *Client) readFromServer() (string, error) {
	return bufio.NewReader(c.Conn).ReadString('\n')
}

// HandleServerCommunication handles data exchange with the server.
func HandleServerCommunication(instance *RaspberryPiInfo, machineID string, handshakes []*entities.Handshake) error {
	client, err := InitClientConnection()
	if err != nil {
		log.Fatalf("[RSP-PI] Failed to initialize client connection: %s", err.Error())
	}
	// defer client.Conn.Close() linter was complaining, but if we exit we should not need to close connection

	request := entities.TCPCreateRaspberryPIRequest{
		Handshakes:    handshakes,
		Jwt:           *instance.JWT,
		MachineID:     machineID,
		EncryptionKey: "",
	}

	wrote, err := client.writeToServerHandshake(request)
	if err != nil {
		return fmt.Errorf("[RSP-PI] Failed to write to server: %s", err.Error())
	}
	log.Printf("[RSP-PI] Wrote %v bytes", wrote)

	response, err := client.readFromServer()
	if err != nil {
		return fmt.Errorf("[RSP-PI] Failed to write to server: %s", err.Error())
	}

	log.Println("[RSP-PI] Response from server:", response)

	return nil
}
