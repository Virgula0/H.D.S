package constants

import (
	"os"
	"path/filepath"
)

const TempDir = "/tmp/hds"
const MachineIDFile = "/etc/machine-id"

var (
	TempPCAPStorage    = filepath.Join(TempDir, "downloads")
	TempHashcatFileDir = filepath.Join(TempDir, "converted")

	PCAPExtension    = ".pcap"
	HashcatExtension = ".hashcat"

	GrpcURL     = os.Getenv("GRPC_URL")
	MachineName = os.Getenv("HOSTNAME")
)

var ListOfDirToCreate = []string{TempPCAPStorage, TempHashcatFileDir}

const (
	CrackStatus     = "cracked"
	ErrorStatus     = "error"
	ExhaustedStatus = "exhausted"
	WorkingStatus   = "working"
)

const (
	HashcatFile   = "hashcatFile"
	HashcatStatus = "status"
	PCAPFile      = "pcapFile"
)
