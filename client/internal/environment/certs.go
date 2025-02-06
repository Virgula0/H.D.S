package environment

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/enums"
)

// containsClientAuth checks if a certificate is for client authentication
func containsClientAuth(usages []x509.ExtKeyUsage) bool {
	for _, usage := range usages {
		if usage == x509.ExtKeyUsageClientAuth {
			return true
		}
	}
	return false
}

// ClassifyFile categorizes a PEM file as caCert, clientCert, or clientKey
func ClassifyFile(fileName string, data []byte) (enums.FileType, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return enums.Unknown, fmt.Errorf("file %s does not contain a valid PEM block", fileName)
	}

	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return enums.Unknown, fmt.Errorf("failed to parse certificate in file %s: %v", fileName, err)
		}

		switch {
		case cert.IsCA:
			return enums.CaCert, nil
		case containsClientAuth(cert.ExtKeyUsage):
			return enums.ClientCert, nil
		default:
			return enums.Unknown, nil
		}

	case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
		return enums.ClientKey, nil

	default:
		return enums.Unknown, nil
	}
}
