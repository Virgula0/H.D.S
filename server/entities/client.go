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
}

type ReturnClientsInstalledResponse struct {
	Length  int       `json:"length"`
	Clients []*Client `json:"clients"`
}

type DeleteClientRequest struct {
	ClientUUID string `json:"client_id" validate:"required"`
}

type DeleteClientResponse struct {
	Status bool `json:"status"`
}
