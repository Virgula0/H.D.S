package entities

const RaspberryPiTableName = "raspberry_pi"

type RaspberryPI struct {
	UserUUID        string `db:"UUID_USER"`
	RaspberryPIUUID string `db:"UUID"`
	MachineID       string `db:"MACHINE_ID"`
	EncryptionKey   string `db:"ENCRYPTION_KEY"`
}
