package main

import (
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/cmd"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/constants"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/daemon"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/utils"
	internalWIFI "github.com/Virgula0/progetto-dp/raspberrypi/internal/wifi"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/wpaparser"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

// initializeInstance sets up the Raspberry Pi instance.
func initializeInstance() (_ *daemon.RaspberryPiInfo, machineID string) {
	machineID, err := utils.MachineID()
	if err != nil {
		log.Fatalf("[RSP-PI] Failed to get machine ID: %s", err.Error())
	}

	// Parse credentials from command line arguments using cobra
	credentials, err := cmd.AuthCommand()
	if err != nil {
		log.Fatalf("[RSP-PI] Failed to get credentials: %s", err.Error())
	}

	return &daemon.RaspberryPiInfo{
		JWT:         new(string),
		FirstLogin:  make(chan bool, 1),
		Credentials: credentials,
	}, machineID
}

// processHandshakes processes Wi-Fi handshakes and prepares them for transmission.
func processHandshakes(env daemon.Environment) []*entities.Handshake {
	handles, err := env.LoadEnvironment()
	if err != nil {
		log.Fatalf("[RSP-PI] Failed to load environment: %s", err.Error())
	}

	handshakes := wpaparser.GetWPA(handles)
	toSend := make([]*entities.Handshake, 0)

	log.Println(strings.Repeat("-", 43))
	for _, handshakeInfo := range handshakes {
		log.Println(*handshakeInfo)

		readContent, err := utils.ReadFileBytes(handshakeInfo.FilePath)
		if err != nil {
			log.Warnf("[RSP-PI] Unable to read '%s': %s", handshakeInfo.FilePath, err.Error())
			continue
		}

		content := utils.BytesToBase64String(readContent)
		toSend = append(toSend, &entities.Handshake{
			SSID:          handshakeInfo.SSID,
			BSSID:         handshakeInfo.BSSID,
			HandshakePCAP: &content,
		})
	}
	log.Println(strings.Repeat("-", 43))
	return toSend
}

// main orchestrates the Raspberry Pi client application.
func main() {
	ticker := time.NewTicker(5 * time.Minute)

	// If it is not a test let's check for connection.
	// This is because we're inside a container we can skip overcomplicating
	if !constants.Test {
		go func() {
			if err := internalWIFI.MonitorWiFiConnection(constants.HomeWIFISSID); err != nil {
				log.Fatal(err.Error())
			}
		}()
	}

	instance, machineID := initializeInstance()
	go instance.Authenticator()

	<-instance.FirstLogin

	env, err := daemon.ChooseEnvironment()
	if err != nil {
		log.Fatalf("[RSP-PI] Failed to choose environment: %s", err.Error())
	}

	for {
		handshakes := processHandshakes(env)
		err := daemon.HandleServerCommunication(instance, machineID, handshakes)
		if err != nil {
			log.Fatal(err.Error())
		}
		<-ticker.C
	}
}
