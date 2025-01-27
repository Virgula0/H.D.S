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
	"golang.org/x/crypto/bcrypt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

var readMessageMutex sync.Mutex
var writeMessageMutex sync.Mutex

// readMessageSize the first message from the client is the length of the content will be sent so we can initialize a buffer
func (wr *TCPServer) readMessageSize(reader *bufio.Reader) (int64, error) {
	readMessageMutex.Lock()
	defer readMessageMutex.Unlock()
	lengthOfMessage, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	lengthOfMessage = strings.TrimSpace(lengthOfMessage)
	return strconv.ParseInt(lengthOfMessage, 10, 64)
}

// readMessageContent read the real message from the client
func (wr *TCPServer) readMessageContent(reader *bufio.Reader, size int64) ([]byte, error) {
	readMessageMutex.Lock()
	defer readMessageMutex.Unlock()
	buffer := make([]byte, size)
	_, err := io.ReadFull(reader, buffer)
	return buffer, err
}

// writeErrorToClient refactored function to send error whenever happens to the client
func (wr *TCPServer) writeErrorToClient(client net.Conn, message string) {
	writeMessageMutex.Lock()
	defer writeMessageMutex.Unlock()
	_, err := client.Write([]byte(message + "\n"))
	if err != nil {
		log.Errorf("[TCP/IP] Error writing to client: %s", err.Error())
	}
}

func (wr *TCPServer) processLoginMessage(buffer []byte, client net.Conn) error {
	var loginRequest entities.AuthRequest

	// Unmarshal the request
	if err := json.Unmarshal(buffer, &loginRequest); err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Invalid request format: %s", err.Error()))
		return err
	}

	// Validate the request
	if err := utils.ValidateGenericStruct(loginRequest); err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Invalid request data: %s", err.Error()))
		return err
	}

	user, role, err := wr.usecase.GetUserByUsername(loginRequest.Username)

	if err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("login failed: %s", err.Error()))
		return err
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("login failed: %s", customErrors.ErrInvalidCredentials))
		return fmt.Errorf("login failed, wrong password for user %s", user.Username)
	}

	// Create the auth token
	token, err := wr.usecase.CreateAuthToken(user.UserUUID, role.RoleString)
	if err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("login failed: %s", customErrors.ErrInvalidCredentials))
		return err
	}

	_, err = client.Write([]byte(token + "\n"))
	return err
}

// processHandshakeMessage performs main tcp server actions
func (wr *TCPServer) processHandshakeMessage(buffer []byte, client net.Conn) error {
	var createRequest TCPCreateRaspberryPIRequest

	// Unmarshal the request
	if err := json.Unmarshal(buffer, &createRequest); err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Invalid request format: %s", err.Error()))
		return err
	}

	// Validate the request
	if err := utils.ValidateGenericStruct(createRequest); err != nil {
		wr.writeErrorToClient(client, fmt.Sprintf("Invalid request data: %s", err.Error()))
		return err
	}

	// Create Raspberry PI
	_, err := wr.createRaspberryPI(&createRequest)
	if err != nil {
		errParsed := wr.handleCreationError(err, client)
		if errParsed != nil {
			wr.writeErrorToClient(client, errParsed.Error())
		}
	}

	// Process Handshakes
	handshakeSavedIDs, err := wr.processHandshakes(createRequest)
	if err != nil {
		wr.writeErrorToClient(client, err.Error())
		return err
	}

	// Send response
	response := strings.Join(handshakeSavedIDs, ";") + "\n"
	_, err = client.Write([]byte(response))
	return err
}

// processHandshakes read handshake data from request and saves it into the database
func (wr *TCPServer) processHandshakes(request TCPCreateRaspberryPIRequest) ([]string, error) {
	handshakeSavedIDs := make([]string, 0)
	for _, handshake := range request.Handshakes {
		if handshake.BSSID == "" || handshake.SSID == "" {
			continue
		}
		handshakeID, err := wr.createHandshake(request.Jwt, handshake)
		if err != nil {
			if errors.Is(err, customErrors.ErrHandshakeAlreadyPresent) {
				return nil, err
			}
			return nil, fmt.Errorf("error creating handshake: %l", err)
		}
		handshakeSavedIDs = append(handshakeSavedIDs, handshakeID)
	}
	if len(handshakeSavedIDs) == 0 {
		return nil, fmt.Errorf("no valid handshakes provided")
	}
	return handshakeSavedIDs, nil
}

// handleCreationError useful function for handling the error returned from createRaspberryPI. If the error is a duplicate error
// we can ignore it, as we assume the device already exists
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
			wr.writeErrorToClient(client, errWrite.Error())
		}
	}
	return err
}

// createRaspberryPI create a raspberrypi entity in the database if it does not exist
func (wr *TCPServer) createRaspberryPI(request *TCPCreateRaspberryPIRequest) (result []byte, err error) {

	data, err := wr.usecase.GetDataFromToken(request.Jwt)

	if err != nil {
		return nil, err
	}

	userID := data[constants.UserIDKey].(string)

	raspID, err := wr.usecase.CreateRaspberryPI(userID, request.MachineID, request.EncryptionKey)

	return []byte(raspID), err
}

// createHandshake create a new handshake if it does not exist
func (wr *TCPServer) createHandshake(jwt string, handshake *entities.Handshake) (result string, err error) {

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
		return "", customErrors.ErrHandshakeAlreadyPresent
	}

	// TODO: use encryption key of the raspberryPI for exchanging handshakes bytes securely
	handshakeID, err := wr.usecase.CreateHandshake(userID, handshake.SSID, handshake.BSSID, constants.NothingStatus, *handshake.HandshakePCAP)

	if err != nil {
		return "", err
	}

	return handshakeID, err
}
