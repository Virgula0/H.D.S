package constants

import (
	"os"
)

const PCAPExtension = ".pcap"
const MachineIDFile = "/etc/machine-id"

var (
	ServerHost = os.Getenv("SERVER_HOST")
	ServerPort = os.Getenv("SERVER_PORT")

	TCPAddress = os.Getenv("TCP_ADDRESS")
	TCPPort    = os.Getenv("TCP_PORT")

	Test = os.Getenv("TEST") == "True"

	HomeWIFISSID = os.Getenv("HOME_WIFI")
)
