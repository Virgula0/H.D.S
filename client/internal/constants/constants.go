package constants

import (
	"os"
	"path/filepath"
)

const TempDir = "/tmp/hds"
const MachineIDFile = "/etc/machine-id"
const HostnameFile = "/etc/hostname"
const FileToCrackPlaceHolder = "FILE_TO_CRACK"

var (
	TempPCAPStorage    = filepath.Join(TempDir, "downloads")
	TempHashcatFileDir = filepath.Join(TempDir, "converted")

	PCAPExtension    = ".pcap"
	HashcatExtension = ".hashcat"

	GrpcURL = os.Getenv("GRPC_URL")
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
