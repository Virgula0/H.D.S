package raspberrypi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
)

func (wr *TCPServer) readMessageSize(reader *bufio.Reader) (int64, error) {
	lengthOfMessage, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	lengthOfMessage = strings.TrimSpace(lengthOfMessage)
	return strconv.ParseInt(lengthOfMessage, 10, 64)
}

func (wr *TCPServer) readMessageContent(reader *bufio.Reader, size int64) ([]byte, error) {
	buffer := make([]byte, size)
	_, err := io.ReadFull(reader, buffer)
	return buffer, err
}

func (wr *TCPServer) processMessage(buffer []byte, client net.Conn) error {
	var createRequest TCPCreateRaspberryPIRequest

	// Unmarshal the request
	if err := json.Unmarshal(buffer, &createRequest); err != nil {
		wr.writeError(client, "Invalid request format: "+err.Error())
		return err
	}

	// Validate the request
	if err := utils.ValidateGenericStruct(createRequest); err != nil {
		wr.writeError(client, "Invalid request data: "+err.Error())
		return err
	}

	// Create Raspberry PI
	_, err := wr.CreateRaspberryPI(&createRequest)
	if err != nil {
		errParsed := wr.handleCreationError(err, client)
		if errParsed != nil {
			wr.writeError(client, errParsed.Error())
		}
	}

	// Process Handshakes
	handshakeSavedIDs, err := wr.processHandshakes(createRequest)
	if err != nil {
		wr.writeError(client, err.Error())
		return err
	}

	// Send response
	response := strings.Join(handshakeSavedIDs, ";") + "\n"
	_, err = client.Write([]byte(response))
	return err
}

func (wr *TCPServer) processHandshakes(request TCPCreateRaspberryPIRequest) ([]string, error) {
	handshakeSavedIDs := make([]string, 0)
	for _, handshake := range request.Handshakes {
		if handshake.BSSID == "" || handshake.SSID == "" {
			continue
		}
		handshakeID, err := wr.CreateHandshake(request.Jwt, handshake)
		if err != nil {
			return nil, fmt.Errorf("error creating handshake: %w", err)
		}
		handshakeSavedIDs = append(handshakeSavedIDs, handshakeID)
	}
	if len(handshakeSavedIDs) == 0 {
		return nil, fmt.Errorf("no valid handshakes provided")
	}
	return handshakeSavedIDs, nil
}

func (wr *TCPServer) handleCreationError(err error, client net.Conn) error {
	var mysqlErr *mysql.MySQLError
	switch {
	case errors.As(err, &mysqlErr) && customErrors.ErrCodeDuplicateEntry == mysqlErr.Number:
		log.Warn("[TCP/IP] RaspberryPI already exists")
		err = nil
	default:
		log.Errorf("[TCP/IP] Error creating RaspberryPI: %s", err.Error())
		_, errWrite := client.Write([]byte(err.Error() + "\n"))
		if errWrite != nil {
			wr.writeError(client, errWrite.Error())
		}
	}
	return err
}

func (wr *TCPServer) writeError(client net.Conn, message string) {
	_, err := client.Write([]byte(message + "\n"))
	if err != nil {
		log.Errorf("[TCP/IP] Error writing to client: %s", err.Error())
	}
}

func (wr *TCPServer) CreateRaspberryPI(request *TCPCreateRaspberryPIRequest) (result []byte, err error) {

	data, err := wr.usecase.GetDataFromToken(request.Jwt)

	if err != nil {
		return nil, err
	}

	userID := data[constants.UserIDKey].(string)

	raspID, err := wr.usecase.CreateRaspberryPI(userID, request.MachineID, request.EncryptionKey)

	return []byte(raspID), err
}

func (wr *TCPServer) CreateHandshake(jwt string, handshake *entities.Handshake) (result string, err error) {

	data, err := wr.usecase.GetDataFromToken(jwt)

	if err != nil {
		return "", err
	}

	userID := data[constants.UserIDKey].(string)

	_, saved, err := wr.usecase.GetHandshakesByBSSIDAndSSID(userID, handshake.BSSID, handshake.SSID)

	if err != nil {
		return "", err
	}

	if saved > 0 { // we don't save the handshake if already saved
		return "", fmt.Errorf("handshake already present")
	}

	// TODO: use encryption key of the raspberryPI for exchanging handshakes bytes securely
	handshakeID, err := wr.usecase.CreateHandshake(userID, handshake.SSID, handshake.BSSID, constants.NothingStatus, *handshake.HandshakePCAP)

	if err != nil {
		return "", err
	}

	return handshakeID, err
}
