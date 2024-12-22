package utils

import (
	"crypto/md5" // #nosec G501
	"encoding/base64"
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/constants"
	"os"
	"path/filepath"
)

func BytesToBase64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// DirExists exists returns whether the given file or directory exists
func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadFileNamesByExtension returns full paths of files with a specific extension in a folder and its subfolders
func ReadFileNamesByExtension(root, extension string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle errors accessing a path
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == extension {
			// Build the full path relative to root
			relativePath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			fullPath := filepath.Join(root, relativePath)
			files = append(files, fullPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
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

func MachineID() (string, error) {
	bytes, err := ReadFileBytes(constants.MachineIDFile)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(bytes)), nil // #nosec G401
}
