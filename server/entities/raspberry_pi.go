package entities

const RaspberryPiTableName = "raspberry_pi"

type RaspberryPI struct {
	UserUUID        string `db:"UUID_USER"`
	RaspberryPIUUID string `db:"UUID"`
	MachineID       string `db:"MACHINE_ID"`
	EncryptionKey   string `db:"ENCRYPTION_KEY"`
}

type ReturnRaspberryPiDevicesResponse struct {
	Length  int                          `json:"length"`
	Devices []*CustomRaspberryPIResponse `json:"devices"`
}

// Needed to avoid to display encryption key
type CustomRaspberryPIResponse struct {
	UserUUID        string
	RaspberryPIUUID string
	MachineID       string
}
