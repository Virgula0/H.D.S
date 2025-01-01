package main

import (
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/internal/mygocat"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	// Initialize application environment
	if _, err := environment.InitEnvironment(); err != nil {
		log.Fatal(err)
	}

	// Initialize gRPC client
	client, err := grpcclient.InitClient()

	if err != nil {
		log.Fatal(err)
	}

	// Initialize GUI login window; if exit is true, terminate the application
	if exit := gui.InitLoginWindow(client); exit {
		os.Exit(1)
	}

	defer client.ClientCloser()

	//nolint:gocritic
	/*
		// TODO: fix graphics in another PR
		go func() {
			if exit := gui.RunGUI(gui.StateUpdateCh); exit {
				os.Exit(1)
			}
		}()
	*/

	// Run the authenticator in the background (renew JWT tokens, etc.)
	go client.Authenticator()

	// Gather machine and hostname info
	machineID, err := utils.MachineID()
	if err != nil {
		log.Panic(err.Error())
	}

	// Read hostname, it is useful for giving to the client a human-readable name
	hostnameBytes, err := utils.ReadFileBytes(constants.HostnameFile)
	if err != nil {
		log.Errorf("Unable to read hostname file: %s", err.Error())
	}
	hostname := string(hostnameBytes)

	// Retrieve client info from server
	info, err := client.GetClientInfo(hostname, machineID)
	if err != nil {
		log.Panic(err.Error())
	}
	clientUUID := info.GetClientUuid()

	// Open the HashcatChat stream
	stream, err := client.HashcatChat()
	if err != nil {
		log.Panic(err.Error())
	}

	// Continuously listen for new tasks
	for {
		if err := mygocat.ListenForHashcatTasks(stream, client, clientUUID); err != nil {
			log.Errorf("[CLIENT] Connection closed or error occurred: %s", err.Error())
			return
		}
	}
}
