package entities

const HandshakeTableName = "handshake"

type Handshake struct {
	UserUUID         string `db:"UUID_USER"`
	ClientUUID       string `db:"UUID_ASSIGNED_CLIENT"`
	RaspberryPIUUID  string `db:"UUID_ASSIGNED_RASPBERRY_PI"`
	UUID             string `db:"UUID"`
	SSID             string `db:"SSID"`
	BSSID            string `db:"BSSID"`
	UploadedDate     string `db:"UPLOADED_DATE"`
	Status           string `db:"STATUS"`
	CrackedDate      string `db:"CRACKED_DATE"`
	HashcatOptions   string `db:"HASHCAT_OPTIONS"`
	HashcatLogs      string `db:"HASHCAT_LOGS"`
	CrackedHandshake string `db:"CRACKED_HANDSHAKE"`
	HandshakePCAP    string `db:"HANDSHAKE_PCAP"`
}
