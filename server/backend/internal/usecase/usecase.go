package usecase

import (
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
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
		return uuid.UUID{}, errors.ErrUnableToGetDataFromToken
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
func (uc *Usecase) UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return uc.repo.UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake)
}

func (uc *Usecase) UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return uc.repo.UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake)
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
