package middlewares

import (
	"context"
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
)

// Middleware function to ensure the token is valid
func (u *TokenAuth) EnsureTokenIsValid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := u.TokenValidation(r, w)
		if token == "" {
			// Not authorized, ResponseWriter already written, no need to call c.JSON again
			return
		}

		// Store the token in the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.TokenConstant, token)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
