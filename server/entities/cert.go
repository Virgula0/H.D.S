package entities

const CertTableName = "certs"

type Cert struct {
	CertUUID   string `db:"UUID"`
	ClientUUID string `db:"CLIENT_UUID"`
	CACert     string `db:"CA_CERT"`
	ClientCert string `db:"CLIENT_CERT"`
	ClientKey  string `db:"CLIENT_KEY"`
}
