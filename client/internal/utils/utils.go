// #nosec G501,G401 // remove md5 warnings
package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/enums"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DirOrFileExists exists returns whether the given file or directory exists
func DirOrFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDirectory creates a directory at the given fullPath.
// It also creates any necessary parent directories.
func CreateDirectory(fullPath string) error {
	exists, err := DirOrFileExists(fullPath)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	err = os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func CreateFileWithBytes(fullPath string, data []byte) error {
	// Ensure the parent directory exists
	dir := filepath.Dir(fullPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create and open the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

func CreateMD5RandomFile(fullPath, ext string, bytes []byte) (string, error) {
	filePath := filepath.Join(fullPath, fmt.Sprintf("%x", md5.Sum([]byte(GenerateToken(20)+time.Now().String())))+ext)
	err := CreateFileWithBytes(filePath, bytes)

	if err != nil {
		return "", err
	}

	return filePath, nil
}

// GenerateToken generates a secure token of the specified length.
func GenerateToken(length int) string {
	// Calculate the required byte length for the token
	byteLength := length / 2

	// Generate random bytes
	bytes := make([]byte, byteLength)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Errorf("cannot create randon token %w", err))
	}

	// Convert the random bytes to a hexadecimal string
	token := hex.EncodeToString(bytes)

	return token
}

// ReadFileBytes returns the bytes of a file given a root path.
func ReadFileBytes(rootPath string) ([]byte, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Keys struct {
	CACert     []byte
	ClientCert []byte
	ClientKey  []byte
}

// CallBackFunc Useful struct for a receiver for passing RecursiveDirectoryWalk to WalkDir maintaining the same function signature
type CallBackFunc struct {
	CallBack func(fileName string, data []byte) (enums.FileType, error)
	Keys     *Keys
}

func (c *CallBackFunc) RecursiveDirectoryWalk(path string, d fs.DirEntry, _ error) error {
	if d.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
		return nil // skip
	}

	fileBytes, err := ReadFileBytes(path)
	fileName := filepath.Base(path)

	if err != nil {
		return err
	}

	typeName, err := c.CallBack(fileName, fileBytes)

	if err != nil {
		return fmt.Errorf("error processing file %s: %v", fileName, err)
	}

	log.Infof("[CLIENT] Recognized crypto file: %s -> Type: %s", fileName, typeName)

	switch typeName {
	case enums.CaCert:
		c.Keys.CACert = fileBytes
	case enums.ClientCert:
		c.Keys.ClientCert = fileBytes
	case enums.ClientKey:
		c.Keys.ClientKey = fileBytes
	default:
		return fmt.Errorf("unknown crypto type: %s", typeName)
	}

	return nil
}

func StringBase64DataToBinary(input string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)

	if err != nil {
		return nil, err
	}
	return decoded, nil

}

func MachineID() (string, error) {
	bytes, err := ReadFileBytes(constants.MachineIDFile)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(bytes)), nil // #nosec G401
}
