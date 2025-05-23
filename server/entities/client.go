package entities

const ClientTableName = "client"

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

type ReturnClientsInstalledResponse struct {
	Length  int       `json:"length"`
	Clients []*Client `json:"clients"`
	Certs   []*Cert   `json:"certs"`
}

type DeleteClientRequest struct {
	ClientUUID string `json:"client_id" validate:"required"`
}

type DeleteClientResponse struct {
	Status bool `json:"status"`
}

type UpdateEncryptionClientStatusRequest struct {
	ClientUUID string `json:"clientUUID" validate:"required,uuid4"`
	Status     *bool  `json:"status" validate:"required"`
}

type UpdateEncryptionClientStatusResponse struct {
	Status bool `json:"status"`
}
