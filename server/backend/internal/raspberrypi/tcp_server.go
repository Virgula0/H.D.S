package raspberrypi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/go-sql-driver/mysql"
	"strings"

	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
)

type TCPHandler interface {
	RunTCPServer()
}

type Wrapper struct {
	cc      net.Conn
	usecase *usecase.Usecase
}

// Read reads data into the provided byte slice from the underlying net.Conn. It returns the number of bytes read and an error.
func (wr *Wrapper) Read(b []byte) (n int, err error) {

	numberOfBytesRead, err := wr.cc.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}

	return numberOfBytesRead, io.EOF
}

type TCPCreateRaspberryPIRequest struct {
	Handshakes    []*entities.Handshake
	Jwt           string `validate:"required"`
	UserUUID      string `validate:"required,len=36"`
	MachineID     string `validate:"required,len=32"`
	EncryptionKey string `validate:"required,len=64"`
}

func (wr *TCPServer) RunTCPServer() error {
	log.Printf("[TCP/IP Server] TCP/IP server running on %s", wr.w.Addr())

	for {
		client, err := wr.w.Accept() // blocking channel

		if err != nil {
			log.Errorf("[TCP/IP] Error accepting connection: %s", err.Error())
			continue // skip
		}

		wrapper := &Wrapper{
			cc:      client,
			usecase: wr.usecase, // inject usecase
		}

		// new goroutine in order to accept new connection after
		go func() {
			defer client.Close()
			read, err := io.ReadAll(wrapper) // calls (wr Wrapper) Read(b []byte). it's a reader after implementation

			if err != nil {
				log.Errorf("[TCP/IP] Error reading from client: %s", err.Error())
				_, errWrite := client.Write([]byte(err.Error() + "\n"))
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
				}
				return
			}

			log.Println("read: " + string(read))

			var createRequest TCPCreateRaspberryPIRequest

			// Unmarshal into the createRequest struct
			err = json.Unmarshal(read, &createRequest)
			if err != nil {
				log.Errorf("[TCP/IP] Cannot unmarshal TCPCreateRaspberryPIRequest from client: %s", err.Error())
				_, errWrite := client.Write([]byte(err.Error() + "\n"))
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
				}
				return // critical can close connection with client
			}
			errValidation := utils.ValidateGenericStruct(createRequest)

			if errValidation != nil {
				log.Errorf("[TCP/IP] Error, request from the client is not valid %s", errValidation.Error())
				_, errWrite := client.Write([]byte(errValidation.Error() + "\n"))
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
				}
				return // critical can close connection with client
			}

			createdID, errCreation := wrapper.CreateRaspberryPI(&createRequest)
			createdID = append(createdID, byte('\n'))

			var mysqlErr *mysql.MySQLError

			switch {
			case errors.As(errCreation, &mysqlErr) && customErrors.ErrCodeDuplicateEntry == mysqlErr.Number:
				// Handle the duplicate entry error
				log.Warn("[TCP/IP] RaspberryPI already exists")

			case errCreation != nil:
				// Handle other errors
				log.Errorf("[TCP/IP] Error, cannot create RaspberryPI %s", errCreation.Error())
				_, errWrite := client.Write([]byte(errCreation.Error() + "\n"))
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
				}
				return // critical can close connection with client
			default:
				_, errWrite := client.Write(createdID)
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
					return
				}
			}

			handshakeSavedIDs := make([]string, 0)
			// proceed with checking if the handshakes exist
			switch {
			case len(createRequest.Handshakes) > 0:
				for _, handshake := range createRequest.Handshakes {
					bssid, essid := handshake.BSSID, handshake.SSID
					if bssid == "" || essid == "" {
						continue // skip
					}
					handshakeID, err := wrapper.CreateHandshake(createRequest.Jwt, string(createdID), handshake)
					if err != nil {
						_, errWrite := client.Write([]byte(errCreation.Error() + "\n"))
						if errWrite != nil {
							log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
							return
						}
					}
					// nope, write will end transmission with the client. need to return an array of ids
					handshakeSavedIDs = append(handshakeSavedIDs, handshakeID)
				}
			default:
				_, errWrite := client.Write([]byte("No handshakes provided\n"))
				if errWrite != nil {
					log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
					return
				}
			}

			_, errWrite := client.Write([]byte(strings.Join(handshakeSavedIDs, ";") + "\n"))
			if errWrite != nil {
				log.Errorf("[TCP/IP] Error, cannot reply to the client %s", errWrite.Error())
				return
			}
		}()
	}
}

func (wr *Wrapper) CreateRaspberryPI(request *TCPCreateRaspberryPIRequest) (result []byte, err error) {

	data, err := wr.usecase.GetDataFromToken(request.Jwt)

	if err != nil {
		return nil, err
	}

	userID := data[constants.UserIDKey].(string)

	raspID, err := wr.usecase.CreateRaspberryPI(userID, request.MachineID, request.EncryptionKey)

	return []byte(raspID), err
}

func (wr *Wrapper) CreateHandshake(jwt, rspID string, handshake *entities.Handshake) (result string, err error) {

	data, err := wr.usecase.GetDataFromToken(jwt)

	if err != nil {
		return "", err
	}

	userID := data[constants.UserIDKey].(string)

	_, saved, err := wr.usecase.GetHandshakesByBSSIDAndSSID(userID, handshake.BSSID, handshake.SSID)

	if saved > 0 { // we don't save the handshake if already saved
		return "", fmt.Errorf("undshake already present")
	}

	// TODO: use encryption key of the raspberryPI for exchanging handshakes bytes securely
	handshakeID, err := wr.usecase.CreateHandshake(userID, rspID, handshake.SSID, handshake.BSSID, constants.NothingStatus, *handshake.HandshakePCAP)

	if err != nil {
		return "", err
	}

	return handshakeID, err
}
