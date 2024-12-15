package authenticate

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	rr "github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (u Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	c := rr.ResponseInitializer{ResponseWriter: w}

	var request AuthRequest

	err := utils.ValidateJSON(&request, r)

	if err != nil {
		c.JSON(http.StatusBadRequest, rr.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	// Check user existence
	user, role, err := u.Usecase.GetUserByUsername(request.Username)
	if err != nil {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, rr.UniformResponse{
			StatusCode: statusCode,
			Details:    err.Error(),
		})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, rr.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrInvalidCreds.Error(),
		})
		return
	}

	// Create the auth token
	token, err := u.Usecase.CreateAuthToken(user.UserUUID, role.RoleString)
	if err != nil {
		statusCode := http.StatusInternalServerError

		c.JSON(statusCode, rr.UniformResponse{
			StatusCode: statusCode,
			Details:    err.Error(),
		})
		return
	}

	// Send the token in response
	c.JSON(http.StatusOK, rr.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    token,
	})
}
