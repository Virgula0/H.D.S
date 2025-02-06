package main

import (
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/gui"
	"github.com/Virgula0/progetto-dp/client/internal/mygocat"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

func invokeHealthCheck(client *grpcclient.Client) {
	log.Warn("[CLIENT] Invoking healthcheck this can take a while... ")
	// ping server with test method
	if _, err := client.Test(); err != nil {
		log.Fatalf("[CLIENT] server seems to be down or unreachable, %v", err)
	}
	log.Info("[CLIENT] Healthcheck done. ")
}

func invokeGUI(client *grpcclient.Client) {
	// Initialize GUI login window; if exit is true, terminate the application
	if exit := gui.InitLoginWindow(client); exit {
		log.Warn("[CLIENT] User exited")
		os.Exit(0)
	}

	// Main process window
	go func() {
		if closed := gui.RunGUI(gui.StateUpdateCh); closed {
			log.Warn("[CLIENT] User exited")
			os.Exit(0)
		}
	}()
}

func invokeGetMachineInfo() (machineID, hostname string, err error) {
	// Gather machine and hostname info
	machineID, err = utils.MachineID()
	if err != nil {
		return "", "", err
	}

	// Read hostname, it is useful for giving to the client a human-readable name
	hostnameBytes, err := utils.ReadFileBytes(constants.HostnameFile)
	if err != nil {
		return "", "", err
	}

	return machineID, string(hostnameBytes), nil
}

func invokeClientStructInit(client *grpcclient.Client, info *pb.GetClientInfoResponse) mygocat.TaskHandler {
	// Fill up client info struct received from server
	client.EntityClient = &entities.Client{
		UserUUID:             info.GetUserUuid(),
		ClientUUID:           info.GetClientUuid(),
		Name:                 info.GetName(),
		LatestIP:             info.GetLatestIp(),
		CreationTime:         info.GetCreationTime(),
		LatestConnectionTime: info.GetLastConnectionTime(),
		MachineID:            info.GetMachineId(),
		EnabledEncryption:    info.GetEnabledEncryption(),
	}

	// Open the HashcatChat stream
	stream, err := client.HashcatChat()
	if err != nil {
		log.Fatalf("[CLIENT] %v", err.Error())
	}

	// Initialize type struct
	return mygocat.TaskHandler{
		Gocat: &mygocat.Gocat{
			Stream: stream,
			Client: client,
		},
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	env, err := environment.InitEnvironment()
	// Initialize application environment
	if err != nil {
		log.Fatalf("[CLIENT] %v", err.Error())
	}

	// Initialize gRPC client
	client, err := grpcclient.InitClient(env)

	if err != nil {
		log.Fatalf("[CLIENT] %v", err.Error())
	}

	invokeHealthCheck(client)

	invokeGUI(client)

	// Run the authenticator in the background (renew JWT tokens, etc.)
	go client.Authenticator()

	machineID, hostname, err := invokeGetMachineInfo()
	if err != nil {
		log.Fatalf("[CLIENT] %v", err.Error())
	}

	// Retrieve client info from server
	info, err := client.GetClientInfo(hostname, machineID)
	if err != nil {
		log.Fatalf("[CLIENT] %v", err.Error())
	}

	log.Infof("[CLIENT] Enabled encryption? (%v)", info.GetEnabledEncryption())

	if info.GetEnabledEncryption() && env.EmptyCerts() {
		log.Fatal("[CLIENT] Encryption is enabled but certs are missing")
	}

	gocat := invokeClientStructInit(client, info)

	defer client.ClientCloser()

	// Continuously listen for new tasks
	for {
		if err := gocat.ListenForHashcatTasks(); err != nil {
			log.Errorf("[CLIENT] Connection closed or error occurred: %s", err.Error())
			return
		}
	}
}
