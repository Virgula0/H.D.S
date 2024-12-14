package register

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/entities"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/google/uuid"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type Request struct {
	Username        string `json:"username" binding:"required" validate:"max=250"`
	Password        string `json:"password" binding:"required" validate:"max=250"`
	PasswordConfirm string `json:"confirmation" binding:"required" validate:"max=250"`
}

func (u Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	c := response.ResponseInitializer{ResponseWriter: w}
	var request Request

	if constants.AllowRegistrations == "" {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrRegistrationNotEnabled.Error(),
		})
		return
	}

	// Validate the request
	if err := utils.ValidateJSON(&request, r); err != nil {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    err.Error(),
		})
		return
	}

	// Validate username and password
	if !utils.IsValidUsername(request.Username) {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrBadPUsernameCriteria.Error(),
		})
		return
	}

	if !utils.IsValidPassword(request.Password) {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrBadPasswordCriteria.Error(),
		})
		return
	}

	// Check if password and confirmation match
	if request.Password != request.PasswordConfirm {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrPaswwordAndConfirmationMismatch.Error(),
		})
		return
	}

	// Create the user entity
	userEntity := &entities.User{
		UserUUID: uuid.New().String(),
		Username: request.Username,
		Password: request.Password,
	}

	// Call the usecase to create the user
	err := u.Usecase.CreateUser(userEntity, constants.USER)
	if err != nil {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrUsernameAlreadyTaken.Error(),
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, response.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "registered",
	})
}
