package grpcserver

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"google.golang.org/grpc/peer"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

// Test is just useful for checking correct gRPC server stage
func (s *ServerContext) Test(_ context.Context, _ *pb.HelloRequest) (*pb.HelloResponse, error) {
	// do usecase stuff
	return &pb.HelloResponse{
		Message: "Hello, World!",
	}, nil
}

// GetClientInfo - Creates a client if it does not exist otherwise returns the current one
func (s *ServerContext) GetClientInfo(ctx context.Context, request *pb.GetClientInfoRequest) (*pb.GetClientInfoResponse, error) {
	jwt := request.GetJwt()
	name := request.GetName()
	machineUUID := request.GetMachineId()
	p, ok := peer.FromContext(ctx)

	if !ok {
		return nil, customErrors.ErrNotValidClientIP
	}

	remoteIP := p.Addr.String()

	data, err := s.Usecase.GetDataFromToken(jwt)

	if err != nil {
		return nil, err
	}

	userID := data[constants.UserIDKey].(string)

	client, err := s.Usecase.GetClientInfo(userID, machineUUID)

	if err != nil {
		// We can create a new client since it does not exist
		if errors.Is(err, customErrors.ErrNoClientFound) {
			newID, errClientCreation := s.Usecase.CreateClient(userID, machineUUID, remoteIP, name)

			if errClientCreation != nil {
				return nil, errClientCreation
			}

			// client created
			return &pb.GetClientInfoResponse{
				IsRegistered:       false,
				UserUuid:           userID,
				ClientUuid:         newID,
				Name:               name,
				LatestIp:           remoteIP,
				CreationTime:       time.Now().Format(constants.DateTimeExample),
				LastConnectionTime: time.Now().Format(constants.DateTimeExample),
				MachineId:          machineUUID,
			}, nil
		}
		// otherwise return the error
		return nil, err
	}

	// if here the client exists and no previous errors have been found from the query
	return &pb.GetClientInfoResponse{
		IsRegistered:       true,
		UserUuid:           client.UserUUID,
		ClientUuid:         client.ClientUUID,
		Name:               client.Name,
		LatestIp:           client.LatestIP,
		CreationTime:       client.CreationTime,
		LastConnectionTime: client.LatestConnectionTime,
		MachineId:          client.MachineID,
	}, nil
}

func (s *ServerContext) sendTasksToClients(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	errChannel := make(chan error, 1)
	go func() {
		for {
			// Example of the server sending a message to the client
			handshakes, _, err := s.Usecase.GetHandshakesByStatus(constants.PendingStatus)

			if err != nil {
				errChannel <- fmt.Errorf("[GRPC]: HashcatChat GetHandshakesByStatus -> %s", err.Error())
				return
			}

			tasks := make([]*pb.ClientTask, 0)
			for _, handshake := range handshakes {
				tasks = append(tasks, &pb.ClientTask{
					StartCracking:  true,
					UserId:         handshake.UserUUID,
					ClientUuid:     *handshake.ClientUUID,
					HandshakeUuid:  handshake.UUID,
					HashcatOptions: *handshake.HashcatOptions,
					HashcatPcap:    *handshake.HandshakePCAP,
				})
			}

			err = stream.Send(&pb.ClientTaskMessageFromServer{
				Tasks: tasks,
			})

			if err != nil {
				errChannel <- fmt.Errorf("[GRPC]: HashcatChat -> error sending message: %w", err)
				return // this will end the goroutine if sending fails
			}
			time.Sleep(1 * time.Second) // Add a delay to avoid flooding the client
		}
	}()
	err := <-errChannel
	return err
}

func (s *ServerContext) listenToTasksFromClient(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	// Server listens for client messages.
	for {
		msg, err := stream.Recv() // msg: ClientTaskMessageFromClient
		if err != nil {
			// Handle the error
			if status.Code(err) == codes.Canceled {
				return status.Errorf(codes.NotFound, "Client has closed the connection")
			}
			return status.Errorf(codes.Unknown, "Failed to receive message: %v", err)
		}
		// Process the received message, extract userID
		log.Println("Received from client:")
		data, err := s.Usecase.GetDataFromToken(msg.GetJwt())

		if err != nil {
			return err
		}

		userID := data[constants.UserIDKey].(string)

		handshake, err := s.Usecase.UpdateClientTask(userID, msg.GetHandshakeUuid(), msg.GetClientUuid(), msg.GetStatus(), msg.GetHashcatOptions(), msg.GetHashcatLogs(), msg.GetCrackedHandshake())

		if err != nil {
			return status.Errorf(codes.Internal, "Cannot update client task: %v", err)
		}

		// answer back the client
		if err := stream.Send(&pb.ClientTaskMessageFromServer{
			Tasks: []*pb.ClientTask{
				{
					StartCracking:  false,
					UserId:         userID,
					ClientUuid:     *handshake.ClientUUID,
					HandshakeUuid:  handshake.UUID,
					HashcatOptions: *handshake.HashcatOptions,
					HashcatPcap:    *handshake.HandshakePCAP,
				},
			},
		}); err != nil {
			return status.Errorf(codes.Internal, "Cannot answer to the client after an update: %v", err)
		}
	}
}

func (s *ServerContext) HashcatTaskChat(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	/*
		Here is the logic for this part:
		- We select from table all tasks with pending state (the user has requested to crack it)
		- We send a message to all clients and if the uuid matches, then the client will reply to the server updating the status
	*/
	err := s.sendTasksToClients(stream)

	if err != nil {
		return err
	}

	/*
		- The client will start the cracking process
		- The client will update the status and hashcat logs once cracking has started
	*/
	err = s.listenToTasksFromClient(stream)

	if err != nil {
		return err
	}

	return nil
}
