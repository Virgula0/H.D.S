package wordlisthandler

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
)

const (
	syncInterval     = 5 * time.Second
	chunkSize        = 4096
	hashAlgorithm    = "%x" // Using MD5 for compatibility
	hiddenFilePrefix = "."
)

var serverList = make(map[string]string, 0)

type Handler struct {
	Handler *environment.ServiceHandler
	Client  *grpcclient.Client
}

func (h *Handler) WordlistSync() error {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Info("[CLIENT] Starting wordlist sync cycle")

		if err := h.syncCycle(); err != nil {
			log.Errorf("[CLIENT] Sync cycle failed: %v", err)
			return err
		}
	}
	return nil
}

func (h *Handler) syncCycle() error {
	response, err := h.Client.GetWordlistInfo()
	if err != nil {
		return fmt.Errorf("failed to get wordlist info: %w", err)
	}

	err = h.syncServerWordlists(response)
	if err != nil {
		return fmt.Errorf("failed to sync server wordlists: %w", err)
	}

	if err := h.uploadNewWordlist(); err != nil {
		return fmt.Errorf("failed to upload new wordlists: %w", err)
	}

	return nil
}

func (h *Handler) syncServerWordlists(response *pb.GetWordlistResponse) error {

	for _, wordlistInfo := range response.GetInfo() {
		wlEntity := &entities.Wordlist{
			UUID:         wordlistInfo.GetWordlistName(),
			WordlistName: wordlistInfo.GetWordlistName(),
			WordlistHash: wordlistInfo.GetWordlistHash(),
			WordlistSize: int(wordlistInfo.GetWordlistSize()),
		}

		err := h.Handler.Usecase.CreateWordlist(wlEntity)
		switch {
		case err == nil:
			log.Infof("[CLIENT] Added new wordlist to local DB: %s", wlEntity.WordlistName)
		case strings.Contains(err.Error(), "UNIQUE constraint failed: wordlist.WORDLIST_HASH"):
			log.Warnf("[CLIENT] Duplicate word list write attempt %s", err.Error())
			continue
		default:
			return fmt.Errorf("database error for %s: %w", wlEntity.WordlistName, err)
		}

		serverList[wordlistInfo.GetWordlistHash()] = wlEntity.WordlistName
	}

	return nil
}

func (h *Handler) uploadNewWordlist() error {
	return filepath.WalkDir(constants.WordlistPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Warnf("[CLIENT] Skipping path due to error: %s - %v", path, err)
			return nil
		}

		if shouldSkipFile(d) {
			return nil
		}

		return h.processWordlistFile(path, serverList)
	})
}

func shouldSkipFile(d fs.DirEntry) bool {
	return d.IsDir() || strings.HasPrefix(d.Name(), hiddenFilePrefix)
}

func (h *Handler) processWordlistFile(path string, serverList map[string]string) error {
	fileBytes, err := utils.ReadFileBytes(path)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	fileName := filepath.Base(path)
	fileHash := fmt.Sprintf(hashAlgorithm, md5.Sum(fileBytes))

	if _, exists := serverList[fileHash]; exists {
		log.Infof("[CLIENT] Skipping existing wordlist from the upload: %s (hash: %s)", fileName, fileHash)
		return nil
	}

	log.Infof("[CLIENT] Uploading new wordlist: %s (hash: %s)", fileName, fileHash)

	// update server list
	serverList[fileHash] = fileName

	return h.streamWordlist(fileName, fileBytes)
}

func (h *Handler) streamWordlist(fileName string, content []byte) error {
	stream, err := h.Client.ClientToServerWordlist()
	if err != nil {
		return fmt.Errorf("stream creation failed: %w", err)
	}

	for offset := 0; offset < len(content); offset += chunkSize {
		end := offset + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunk := &pb.Chunk{
			Content:      content[offset:end],
			ClientUuid:   h.Client.EntityClient.ClientUUID,
			Jwt:          *h.Client.Credentials.JWT,
			WordlistName: fileName,
		}

		if errSend := stream.Send(chunk); errSend != nil {
			return fmt.Errorf("chunk send failed at offset %d: %w", offset, errSend)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("stream closure failed: %w", err)
	}

	log.Infof("[CLIENT] Completed upload for %s. Server response: %s", fileName, response.GetCode().String())
	return nil
}
