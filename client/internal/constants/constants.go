package constants

import (
	"os"
	"path/filepath"
	"strings"
)

const TempDir = "/tmp/hds"
const MachineIDFile = "/etc/machine-id"
const HostnameFile = "/etc/hostname"
const FileToCrackPlaceHolder = "FILE_TO_CRACK"
const CertFileDir = "certs"
const WordlistPath = "wordlists"

var MaxGRPCFileSize = 295 << 20 // 295Mb

var (
	TempPCAPStorage    = filepath.Join(TempDir, "downloads")
	TempHashcatFileDir = filepath.Join(TempDir, "converted")

	PCAPExtension    = ".pcap"
	HashcatExtension = ".hashcat"

	GrpcURL     = os.Getenv("GRPC_URL")
	GrpcTimeout = os.Getenv("GRPC_TIMEOUT")

	WipeTables = strings.ToLower(os.Getenv("RESET")) == "true"

	// Database path

	DBPath = func() string {
		home := os.Getenv("HOME")
		_, err := os.ReadDir(home)

		path, errPath := filepath.Abs(home)
		if errPath != nil || err != nil || home == "" {
			return "database.sqlite"
		}

		return filepath.Join(path, ".HDS", "database.sqlite")
	}()
)

var ListOfDirToCreate = []string{TempPCAPStorage, TempHashcatFileDir}

const (
	CrackStatus     = "cracked"
	ErrorStatus     = "error"
	ExhaustedStatus = "exhausted"
	WorkingStatus   = "working"
	PendingStatus   = "pending"
)

const (
	HashcatFile   = "hashcatFile"
	HashcatStatus = "status"
	PCAPFile      = "pcapFile"
)
