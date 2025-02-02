package enums

// FileType represents the classification of a file
type FileType int

const (
	Unknown FileType = iota
	CaCert
	ClientCert
	ClientKey
)

// String method to convert FileType to readable text
func (ft FileType) String() string {
	return [...]string{"Unknown", "CaCert", "ClientCert", "ClientKey"}[ft]
}
