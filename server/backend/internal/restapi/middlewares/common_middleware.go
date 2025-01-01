package middlewares

import (
	"github.com/Virgula0/progetto-dp/server/entities"
	"net/http"
	"strings"

	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
)

type TokenAuth struct {
	Usecase *usecase.Usecase
}

// TokenValidation a refactored function used by auth_middleware
func (u *TokenAuth) TokenValidation(r *http.Request, w http.ResponseWriter) string {
	// Extract the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		ResponseWithError(w, http.StatusBadRequest, "Authorization token required")
		return ""
	}

	// Split the header value to validate the format "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ResponseWithError(w, http.StatusBadRequest, "Invalid authorization header format")
		return ""
	}

	token := parts[1]

	// Check if the token is a valid JWT
	if !utils.IsJWT(token) {
		ResponseWithError(w, http.StatusUnauthorized, "Invalid token: token is not in valid JWT format")
		return ""
	}

	// Validate the token using the Usecase
	isValid, err := u.Usecase.ValidateToken(token)
	if err != nil || !isValid {
		ResponseWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return ""
	}

	// If the token is valid, return it
	return token
}

// ResponseWithError Helper function to respond with an error message
func ResponseWithError(w http.ResponseWriter, statusCode int, message string) {
	c := response.Initializer{ResponseWriter: w}
	c.JSON(statusCode, entities.UniformResponse{
		StatusCode: statusCode,
		Details:    message,
	})
}
