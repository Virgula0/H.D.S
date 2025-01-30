package usecase

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"math/big"
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Usecase struct {
	repo *repository.Repository
}

var blacklistedTokens = make(map[string]bool)

// NewUsecase Dep. injection for usecase. Injecting db -> repo -> usecase
func NewUsecase(repo *repository.Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

// CreateServerCerts InjectCerts Injects generated certs into Repository
func (uc *Usecase) CreateServerCerts() error {
	caCert, caKey, err := uc.createCA()
	if err != nil {
		return err
	}

	// sign server certs
	serverCert, serverKey, err := uc.SignCert(caCert, caKey, utils.GenerateToken(32)) // the clientID is not important for the server, we can generate a random va
	if err != nil {
		return err
	}

	uc.repo.InjectCerts(caCert, caKey, serverCert, serverKey)

	return nil
}

// GetServerCerts if CreateServerCerts has been called before this, no error will be returned
func (uc *Usecase) GetServerCerts() (caCert, caKey, serverCert, serverKey []byte, err error) {
	return uc.repo.GetCerts()
}

func (uc *Usecase) createCA() ([]byte, []byte, error) {
	// Generate a private key for the CA
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("%s %v", customErrors.ErrFailToGeneratePrivateKey, err)
	}

	// Create a CA certificate template
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{constants.OrganizationCertName},
			CommonName:   constants.CertCommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	// Self-sign the CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %v", err)
	}

	marshalPrivateKey, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key error: %v", err)
	}

	// Encode the CA certificate and private key to PEM
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: marshalPrivateKey})

	return caCertPEM, caKeyPEM, nil
}

func (uc *Usecase) SignCert(caCertPEM, caKeyPEM []byte, commonNameClientUUID string) ([]byte, []byte, error) {
	// Decode CA certificate
	certBlock, _ := pem.Decode(caCertPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	// Decode CA key
	keyBlock, _ := pem.Decode(caKeyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key")
	}

	caKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key: %v", err)
	}

	// Generate a private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Create a certificate template
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{constants.OrganizationCertName},
			CommonName:   commonNameClientUUID,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Sign the certificate with the CA
	certDER, err := x509.CreateCertificate(rand.Reader, cert, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	marshalPrivateKey, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key error: %v", err)
	}

	// Encode the certificate and private key to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: marshalPrivateKey})

	return certPEM, keyPEM, nil
}

func (uc *Usecase) GetDataFromToken(tokenInput string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenInput, claims, func(token *jwt.Token) (any, error) {
		return constants.JwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func (uc *Usecase) GetUserIDFromToken(r *http.Request) (uuid.UUID, error) {
	ctx := r.Context()
	token, ok := ctx.Value(constants.TokenConstant).(string)

	if !ok {
		return uuid.UUID{}, customErrors.ErrUnableToGetDataFromToken
	}

	data, err := uc.GetDataFromToken(token)

	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.Parse(data[constants.UserIDKey].(string))
}

func (uc *Usecase) InvalidateToken(token string) {
	blacklistedTokens[token] = true
}

func (uc *Usecase) ValidateToken(tokenInput string) (bool, error) {
	if _, ok := blacklistedTokens[tokenInput]; ok {
		return false, nil
	}

	token, err := jwt.ParseWithClaims(tokenInput, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return constants.JwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return false, err
	}

	return true, nil
}

func (uc *Usecase) CreateAuthToken(userID, role string) (string, error) {
	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"uuid":   (uuid.New()).String(),                 // jwtToken uuid
		"exp":    time.Now().Add(time.Hour * 12).Unix(), // Token expires after 12 hours
	})

	// Sign the token with a secret key
	return token.SignedString(constants.JwtSecretKey)
}

func (uc *Usecase) GetUserByUsername(username string) (*entities.User, *entities.Role, error) {
	return uc.repo.GetUserByUsername(username)
}

func (uc *Usecase) CreateUser(userEntity *entities.User, role constants.Role) error {
	return uc.repo.CreateUser(userEntity, role)
}

func (uc *Usecase) GetClientsInstalled(userUUID string, offset uint) ([]*entities.Client, int, error) {
	return uc.repo.GetClientsInstalledByUserID(userUUID, offset)
}

func (uc *Usecase) CreateClient(userUUID, machineID, latestIP, name string) (string, error) {
	return uc.repo.CreateClient(userUUID, machineID, latestIP, name)
}

func (uc *Usecase) CreateCertForClient(clientUUID string, clientCert, clientKey []byte) (string, error) {
	return uc.repo.CreateCertForClient(clientUUID, clientCert, clientKey)
}

func (uc *Usecase) CreateHandshake(userUUID, ssid, bssid, status, handshakePcap string) (string, error) {
	return uc.repo.CreateHandshake(userUUID, ssid, bssid, status, handshakePcap)
}

func (uc *Usecase) GetRaspberryPI(userUUID string, offset uint) ([]*entities.RaspberryPI, int, error) {
	return uc.repo.GetRaspberryPiByUserID(userUUID, offset)
}

func (uc *Usecase) CreateRaspberryPI(userUUID, machineID, encryptionKey string) (string, error) {
	return uc.repo.CreateRaspberryPI(userUUID, machineID, encryptionKey)
}

func (uc *Usecase) GetHandshakes(userUUID string, offset uint) ([]*entities.Handshake, int, error) {
	return uc.repo.GetHandshakesByUserID(userUUID, offset)
}

func (uc *Usecase) GetClientInfo(userUUID, machineID string) (*entities.Client, error) {
	return uc.repo.GetClientInfo(userUUID, machineID)
}

func (uc *Usecase) GetHandshakesByStatus(filterStatus string) (handshakes []*entities.Handshake, length int, e error) {
	return uc.repo.GetHandshakesByStatus(filterStatus)
}
func (uc *Usecase) UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return uc.repo.UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake)
}

func (uc *Usecase) UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return uc.repo.UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake)
}

func (uc *Usecase) GetHandshakesByBSSIDAndSSID(userUUID, bssid, ssid string) (handshakes []*entities.Handshake, length int, e error) {
	return uc.repo.GetHandshakesByBSSIDAndSSID(userUUID, bssid, ssid)
}

func (uc *Usecase) DeleteClient(userUUID, clientUUID string) (bool, error) {
	return uc.repo.DeleteClient(userUUID, clientUUID)
}

func (uc *Usecase) DeleteRaspberryPI(userUUID, rspUUID string) (bool, error) {
	return uc.repo.DeleteRaspberryPI(userUUID, rspUUID)
}

func (uc *Usecase) DeleteHandshake(userUUID, handshakeUUID string) (bool, error) {
	return uc.repo.DeleteHandshake(userUUID, handshakeUUID)
}
