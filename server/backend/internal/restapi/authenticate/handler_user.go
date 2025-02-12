package authenticate

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"golang.org/x/crypto/bcrypt"
	"net/http"

	rr "github.com/Virgula0/progetto-dp/server/backend/internal/response"
)

// CheckTokenValidity Used for verifying if the JWT is still valid
func (u Handler) CheckTokenValidity(w http.ResponseWriter, _ *http.Request) {
	c := rr.Initializer{ResponseWriter: w}

	c.JSON(http.StatusOK, entities.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "valid",
	})
}

// UpdateUserPassword update user password
func (u Handler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	c := rr.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.UpdateUserPasswordRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	if request.NewPassword != request.NewPasswordConfirm {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    errors.ErrPasswordConfirmationDoNotMatch.Error(),
		})
		return
	}

	user, err := u.Usecase.GetUserByUserID(userID.String())

	if err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OldPassword))
	if err != nil {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, entities.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrOldPasswordMismatch.Error(),
		})
		return
	}

	if !utils.IsValidPassword(request.NewPassword) {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, entities.UniformResponse{
			StatusCode: statusCode,
			Details:    errors.ErrBadPasswordCriteria.Error(),
		})
		return
	}

	err = u.Usecase.UpdateUserPassword(userID.String(), request.NewPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.UpdateUserPasswordResponse{
		Status: "updated",
	})
}
