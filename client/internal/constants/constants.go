package constants

import (
	"os"
	"path/filepath"
)

const TempDir = "/tmp/hds"
const MachineIDFile = "/etc/machine-id"
const HostnameFile = "/etc/hostname"
const FileToCrackPlaceHolder = "FILE_TO_CRACK"
const CertFileDir = "certs"

var (
	TempPCAPStorage    = filepath.Join(TempDir, "downloads")
	TempHashcatFileDir = filepath.Join(TempDir, "converted")

	PCAPExtension    = ".pcap"
	HashcatExtension = ".hashcat"

	GrpcURL     = os.Getenv("GRPC_URL")
	GrpcTimeout = os.Getenv("GRPC_TIMEOUT")
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
