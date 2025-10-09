package wordlisthandler

import (
	"crypto/md5" //nolint:gosec // allow md5
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/customerrors"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/environment"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	pb "github.com/Virgula0/progetto-dp/client/protobuf/hds"
	log "github.com/sirupsen/logrus"
)

const (
	syncInterval     = 1 * time.Minute
	chunkSize        = 4096
	hashAlgorithm    = "%x" // Using MD5 for compatibility
	hiddenFilePrefix = "."
)

type Handler struct {
	Handler *environment.ServiceHandler
	Client  *grpcclient.Client
}

func (h *Handler) WordlistSync() {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		log.Info("[CLIENT] Starting wordlist sync cycle")

		if err := h.syncCycle(); err != nil {
			log.Errorf("[CLIENT] Error syncing wordlist: %v", err)
		}
		<-ticker.C
	}
}

func (h *Handler) syncCycle() error {
	response, err := h.Client.GetWordlistInfo() // Get server's current wordlist hashes
	if err != nil {
		return fmt.Errorf("failed to get wordlist info: %w", err)
	}

	// Track server-side hashes to avoid uploading duplicates
	serverHashes := make(map[string]bool)
	for _, info := range response.GetInfo() {
		serverHashes[info.GetWordlistHash()] = true
	}

	// Sync server wordlists to client
	if err := h.syncServerWordlists(response); err != nil {
		return fmt.Errorf("failed to sync server wordlists: %w", err)
	}

	// Upload new client wordlists to server
	if err := h.uploadNewWordlist(serverHashes); err != nil {
		return fmt.Errorf("failed to upload new wordlists: %w", err)
	}

	return nil
}

func (h *Handler) syncServerWordlists(response *pb.GetWordlistResponse) error {
	for _, wordlistInfo := range response.GetInfo() {
		serverHash := wordlistInfo.GetWordlistHash()

		// Skip if wordlist already exists locally (by hash)
		_, err := h.Handler.Usecase.GetWordlistByHash(serverHash)
		if err != nil && !errors.Is(err, customerrors.ErrNoRowsFound) {
			return err
		}

		// Add new wordlist to local DB and download
		wlEntity := &entities.Wordlist{
			UUID:                 wordlistInfo.GetWordlistId(),
			UserUUID:             h.Client.EntityClient.UserUUID,
			ClientUUID:           h.Client.EntityClient.ClientUUID,
			WordlistName:         wordlistInfo.GetWordlistName(),
			WordlistHash:         serverHash,
			WordlistSize:         int(wordlistInfo.GetWordlistSize()),
			WordlistLocationPath: wordlistInfo.GetWordlistLocationPath(),
		}

		if err := h.Handler.Usecase.CreateWordlist(wlEntity); err != nil {
			if !strings.Contains(err.Error(), "UNIQUE constraint failed: wordlist.WORDLIST_HASH") {
				return fmt.Errorf("failed to create wordlist: %w", err)
			}
			// If already exists, continue without error.
			continue
		}
		log.Infof("[CLIENT] Added new wordlist to local DB: %s", wlEntity.WordlistName)

		if err := h.streamDownloadWordlist(wlEntity); err != nil {
			return fmt.Errorf("download failed for %s: %w", wlEntity.WordlistName, err)
		}
	}
	return nil
}

func (h *Handler) streamDownloadWordlist(ww *entities.Wordlist) error {
	stream, err := h.Client.ServerToClientWordlist(&pb.DownloadWordlist{
		Jwt:        *h.Client.Credentials.JWT,
		ClientId:   h.Client.EntityClient.ClientUUID,
		WordlistId: ww.UUID,
	})
	if err != nil {
		return err
	}

	log.Infof("[CLIENT] Downloading missing wordlist: %s (hash %s)", ww.WordlistName, ww.WordlistHash)

	buffer := make([]byte, 0)
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		buffer = append(buffer, chunk.GetContent()...)
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	saveName := filepath.Join(constants.WordlistPath, ww.WordlistHash, ww.WordlistName)
	hash := fmt.Sprintf("%x", md5.Sum(buffer)) //nolint:gosec // allow md5

	if ww.WordlistHash != hash {
		return fmt.Errorf("error downloading wordlist, expected hash to be %s but got %s", ww.WordlistHash, hash)
	}

	return utils.CreateFileWithBytes(saveName, buffer)
}

func (h *Handler) uploadNewWordlist(serverHashes map[string]bool) error {
	return filepath.WalkDir(constants.WordlistPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Errorf("[CLIENT] Skipping path due to error: %s - %v", path, err)
			return nil
		}

		if shouldSkipFile(d) {
			return nil
		}

		return h.processWordlistFile(path, serverHashes)
	})
}

func shouldSkipFile(d fs.DirEntry) bool {
	return d.IsDir() || strings.HasPrefix(d.Name(), hiddenFilePrefix)
}

func (h *Handler) processWordlistFile(path string, serverHashes map[string]bool) error {
	fileBytes, err := utils.ReadFileBytes(path)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	fileName := filepath.Base(path)
	fileHash := fmt.Sprintf(hashAlgorithm, md5.Sum(fileBytes)) //nolint:gosec // allow md5

	// Skip if server already has this hash
	if serverHashes[fileHash] {
		log.Warnf("[CLIENT] Skipping upload for %s (hash %s): already exists on server", fileName, fileHash)
		return nil
	}

	// Check local DB to avoid re-uploading in the same sync cycle
	_, err = h.Handler.Usecase.GetWordlistByHash(fileHash)
	if err == nil {
		log.Warnf("[CLIENT] Skipping upload for %s (hash %s): already tracked locally", fileName, fileHash)
		return nil
	}
	if !errors.Is(err, customerrors.ErrNoRowsFound) {
		return fmt.Errorf("database error: %w", err)
	}

	// Add to local DB and upload
	ww := &entities.Wordlist{
		UserUUID:             h.Client.EntityClient.UserUUID,
		ClientUUID:           h.Client.EntityClient.ClientUUID,
		WordlistName:         fileName,
		WordlistHash:         fileHash,
		WordlistSize:         len(fileBytes),
		WordlistLocationPath: constants.WordlistPath,
	}

	if err := h.Handler.Usecase.CreateWordlist(ww); err != nil {
		return fmt.Errorf("failed to create wordlist: %w", err)
	}

	log.Infof("[CLIENT] Uploading new wordlist: %s (hash: %s)", fileName, fileHash)
	return h.streamUploadWordlist(fileName, fileBytes)
}

func (h *Handler) streamUploadWordlist(fileName string, content []byte) error {
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

	hash := fmt.Sprintf("%x", md5.Sum(content)) //nolint:gosec // allow md5

	if response.GetHash() != hash {
		return fmt.Errorf("error uploading wordlist, expected hash to be %s but got %s", hash, response.GetHash())
	}

	log.Infof("[CLIENT] Completed upload for %s. Server response: %s", fileName, response.GetCode().String())
	return nil
}
