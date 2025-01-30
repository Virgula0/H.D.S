//nolint:govet // Disabling vet for grpc format string false positives
package grpcserver

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"google.golang.org/grpc/peer"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/entities"
	pb "github.com/Virgula0/progetto-dp/server/protobuf/hds"
)

// Test is just useful for checking correct gRPC server stage
func (s *ServerContext) Test(_ context.Context, _ *pb.HelloRequest) (*pb.HelloResponse, error) {
	// do usecase stuff
	return &pb.HelloResponse{
		Message: "Hello, World!",
	}, nil
}

// Login implements the behavior for Login gRPC method
func (s *ServerContext) Login(_ context.Context, request *pb.AuthRequest) (*pb.UniformResponse, error) {
	user, role, err := s.Usecase.GetUserByUsername(request.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.GetPassword()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%s", customErrors.ErrInvalidCredentials)
	}

	// Create the auth token
	token, err := s.Usecase.CreateAuthToken(user.UserUUID, role.RoleString)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.UniformResponse{
		Status:  "logged_in",
		Details: token,
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

			caCert, caKey, _, _, err := s.Usecase.GetServerCerts()

			if err != nil {
				return nil, err
			}

			// sign certs
			clientCert, clientKey, err := s.Usecase.SignCert(caCert, caKey, newID)
			if err != nil {
				return nil, err
			}

			_, err = s.Usecase.CreateCertForClient(newID, clientCert, clientKey)
			if err != nil {
				return nil, err
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

func (s *ServerContext) HashcatTaskChat(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	errChannel := make(chan error, 1) // Buffered channel to avoid blocking

	/*
		Here is the logic for this part:
		- We select from table all tasks with pending state (the user has requested to crack it)
		- We send a message to all clients and if the uuid matches, then the client will reply to the server updating the status
	*/
	go func() {
		if err := s.sendTasksToClients(stream); err != nil {
			errChannel <- err
		}
	}()

	/*
		- The client will start the cracking process
		- The client will update the status and hashcat logs once cracking has started
	*/
	if err := s.listenToTasksFromClient(stream); err != nil {
		return err
	}

	return <-errChannel
}

// sendTasksToClients sends all pending tasks to clients, the client will recognize the assignment by its clientID
func (s *ServerContext) sendTasksToClients(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	ticker := time.NewTicker(1 * time.Second) // Do not flood client. Update tasks every second
	defer ticker.Stop()

	for {
		<-ticker.C // blocking channel

		handshakes, err := s.getPendingHandshakes()
		if err != nil {
			return err
		}

		tasks := s.prepareTasks(handshakes)
		if len(tasks) == 0 {
			continue
		}

		if err := s.sendTasksToStream(stream, tasks); err != nil {
			return err
		}

		if err := s.updateTaskStatuses(tasks); err != nil {
			return err
		}
	}
}

// ---------- Helper Functions ----------

// getPendingHandshakes retrieves handshakes with pending status.
func (s *ServerContext) getPendingHandshakes() ([]*entities.Handshake, error) {
	handshakes, _, err := s.Usecase.GetHandshakesByStatus(constants.PendingStatus)
	if err != nil {
		return nil, fmt.Errorf("%s %s", customErrors.ErrGetHandshakeStatus, err.Error())
	}
	return handshakes, nil
}

// prepareTasks converts handshakes into tasks for clients.
func (s *ServerContext) prepareTasks(handshakes []*entities.Handshake) []*pb.ClientTask {
	tasks := make([]*pb.ClientTask, 0)
	for _, handshake := range handshakes {
		if handshake.ClientUUID == nil || handshake.HashcatOptions == nil || handshake.HandshakePCAP == nil {
			log.Errorf(
				"%s Missing RaspberryPIUUID, HashcatOptions, or HandshakePCAP for Handshake HandshakeUUID '%s'. Task skipped.",
				customErrors.ErrGetHandshakeStatus,
				handshake.UUID,
			)
			continue
		}

		tasks = append(tasks, &pb.ClientTask{
			StartCracking:  true,
			UserId:         handshake.UserUUID,
			ClientUuid:     *handshake.ClientUUID,
			HandshakeUuid:  handshake.UUID,
			HashcatOptions: *handshake.HashcatOptions,
			HashcatPcap:    *handshake.HandshakePCAP,
			BSSID:          handshake.BSSID,
			SSID:           handshake.SSID,
		})
	}
	return tasks
}

// sendTasksToStream sends tasks to the gRPC stream.
func (s *ServerContext) sendTasksToStream(stream pb.HDSTemplateService_HashcatTaskChatServer, tasks []*pb.ClientTask) error {
	err := stream.Send(&pb.ClientTaskMessageFromServer{Tasks: tasks})
	if err != nil {
		return fmt.Errorf("%s %v", customErrors.ErrCannotAnswerToClient, err)
	}
	return nil
}

// updateTaskStatuses updates the status of sent tasks.
func (s *ServerContext) updateTaskStatuses(tasks []*pb.ClientTask) error {
	for _, task := range tasks {
		_, err := s.Usecase.UpdateClientTask(
			task.GetUserId(),
			task.GetHandshakeUuid(),
			task.GetClientUuid(),
			constants.WorkingStatus,
			task.GetHashcatOptions(),
			"",
			task.GetHashcatPcap(),
		)
		if err != nil {
			return fmt.Errorf("failed to update task status for handshake '%s': %v", task.GetHandshakeUuid(), err)
		}
	}
	return nil
}

// listenToTasksFromClient updates dynamically the coming information from the client. Useful for fast hashcat logs transmission
func (s *ServerContext) listenToTasksFromClient(stream pb.HDSTemplateService_HashcatTaskChatServer) error {
	for {
		// Receive message from client
		msg, err := stream.Recv()
		if err != nil {
			// check if client has disconnected
			if status.Code(err) == codes.Canceled {
				return status.Errorf(codes.NotFound, customErrors.ErrGRPCClosedConnection.Error())
			}
			return status.Errorf(codes.Unknown, fmt.Sprintf("%s %v", customErrors.ErrGRPCFailedToReceive, err))
		}

		log.Printf("[GRPC]: HashcatChat ->Received from client: %v", msg)

		// Process the received message
		data, err := s.Usecase.GetDataFromToken(msg.GetJwt())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, fmt.Sprintf("%s %v", customErrors.ErrInvalidToken, err))
		}

		userID := data[constants.UserIDKey].(string)
		_, err = s.Usecase.UpdateClientTask(
			userID,
			msg.GetHandshakeUuid(),
			msg.GetClientUuid(),
			msg.GetStatus(),
			msg.GetHashcatOptions(),
			msg.GetHashcatLogs(),
			msg.GetCrackedHandshake(),
		)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("%s %v", customErrors.ErrOnUpdateTask, err))
		}
	}
}
