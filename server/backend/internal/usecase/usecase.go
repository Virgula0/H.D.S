package usecase

import (
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/entities"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Usecase struct {
	repo *repository.Repository
}

var blacklistedTokens = make(map[string]bool)

// Dep. injection for usecase. Injecting db -> repo -> usecase
func NewUsecase(repo *repository.Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (uc *Usecase) GetDataFromToken(tokenInput string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenInput, claims, func(token *jwt.Token) (interface{}, error) {
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

	return uuid.Parse(data["userID"].(string))
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
	return uc.repo.GetClientsInstalled(userUUID, offset)
}

func (uc *Usecase) CreateClient(userUUID, machineID, latestIP, name string) (string, error) {
	return uc.repo.CreateClient(userUUID, machineID, latestIP, name)
}

func (uc *Usecase) GetRaspberryPI(userUUID string, offset uint) ([]*entities.RaspberryPI, int, error) {
	return uc.repo.GetRaspberryPI(userUUID, offset)
}
