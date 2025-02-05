package encryption

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
)

// extractSerialNumber extracts the serial number from a PEM-encoded certificate.
func extractSerialNumber(clientCertPEM []byte) (string, error) {
	// Decode the PEM block
	block, _ := pem.Decode(clientCertPEM)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	// Parse the x509 certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Return the Serial Number
	return cert.Subject.SerialNumber, nil
}

// LoadTLSCredentials loads the CA cert and client key from memory.
func LoadTLSCredentials(caCertPEM, clientKeyPEM, clientCertPEM []byte, addTLS bool) (credentials.TransportCredentials, error) {

	var creds = &tls.Config{
		InsecureSkipVerify: true, //#nosec:G402 // use unsecure connection for first client installation or if security is disabled
		ServerName:         utils.GenerateToken(32),
	}

	clientUUID, err := extractSerialNumber(clientCertPEM)
	if err != nil {
		return nil, err
	}

	if addTLS {

		log.Warn("[CLIENT] Setting up a TLS connection")

		// Load the CA certificate
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCertPEM) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		// Load the client certificate and key
		cert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}

		// Create TransportCredentials using the CA cert and client certificate
		creds = &tls.Config{
			// Ignore verification from the server of CN and SIN otherwise we cannot send clientUUID as ServerName.
			// We need to sent it before the handshake, this skips the verification just client side
			InsecureSkipVerify: true,
			RootCAs:            caCertPool,
			Certificates:       []tls.Certificate{cert},
			ServerName:         clientUUID,
		}
	}

	return credentials.NewTLS(creds), nil
}
