package entities

const HandshakeTableName = "handshake"

// Pointers in stracture is to deal with NULL data binding when parsing the rows while querying
type Handshake struct {
	UserUUID         string  `db:"UUID_USER"`
	ClientUUID       *string `db:"UUID_ASSIGNED_CLIENT"`
	UUID             string  `db:"UUID"`
	SSID             string  `db:"SSID"`
	BSSID            string  `db:"BSSID"`
	UploadedDate     string  `db:"UPLOADED_DATE"`
	Status           string  `db:"STATUS"`
	CrackedDate      *string `db:"CRACKED_DATE"`
	HashcatOptions   *string `db:"HASHCAT_OPTIONS"`
	HashcatLogs      *string `db:"HASHCAT_LOGS"`
	CrackedHandshake *string `db:"CRACKED_HANDSHAKE"`
	HandshakePCAP    *string `db:"HANDSHAKE_PCAP"`
}

type GetHandshakeResponse struct {
	Length     int `json:"length"`
	Handshakes []*Handshake
}

type UpdateHandshakeTaskViaAPIResponse struct {
	Success   bool
	Reason    string
	Handshake *Handshake
}

type UpdateHandshakeTaskViaAPIRequest struct {
	HandshakeUUID      string `json:"handshakeUUID" validate:"required"`
	AssignedClientUUID string `json:"clientUUID" validate:"required"`
	HashcatOptions     string `json:"hashcatOptions" validate:"required"`
}

type DeleteHandshakesRequest struct {
	HandshakeUUID string `json:"handshake_uuid"`
}

type DeleteHandshakesResponse struct {
	Status bool `json:"status" validate:"required"`
}
