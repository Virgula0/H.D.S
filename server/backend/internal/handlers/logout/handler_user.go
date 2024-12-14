package logout

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type Token struct {
	Authorization string `header:"authorization" binding:"required"`
}

func (u Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	c := response.ResponseInitializer{ResponseWriter: w}

	// Get the authorization header
	authHeader := r.Header.Get("Authorization")

	// Split the header into parts (should be "Bearer <token>")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		statusCode := http.StatusBadRequest
		json.NewEncoder(w).Encode(response.UniformResponse{
			StatusCode: statusCode,
			Details:    "Invalid authorization header format",
		})
		return
	}

	// Extract the token
	token := parts[1]

	// Check if the token is a valid JWT
	if !utils.IsJWT(token) {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "Invalid token: token is not in valid JWT format",
		})
		return
	}

	// Validate the token
	isValid, err := u.Usecase.ValidateToken(token)
	if err != nil || !isValid {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, response.UniformResponse{
			StatusCode: statusCode,
			Details:    "Invalid or expired token",
		})
		return
	}

	// Invalidate the token
	u.Usecase.InvalidateToken(token)

	// Respond with success
	c.JSON(http.StatusOK, response.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "Logged out successfully",
	})
}
