package encryption

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/credentials"
)

// LoadTLSCredentials loads the CA cert and client key from memory.
func LoadTLSCredentials(caCertPEM, clientKeyPEM, clientCertPEM []byte) (credentials.TransportCredentials, error) {
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
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	return credentials.NewTLS(tlsConfig), nil
}
