package entities

type Client struct {
	UserUUID             string `db:"UUID_USER"`
	ClientUUID           string `db:"UUID"`
	Name                 string `db:"NAME"`
	LatestIP             string `db:"LATEST_IP"`
	CreationTime         string `db:"CREATION_DATETIME"`
	LatestConnectionTime string `db:"LATEST_CONNECTION"`
	MachineID            string `db:"MACHINE_ID"`
	EnabledEncryption    bool   `db:"ENABLED_ENCRYPTION"`
}
