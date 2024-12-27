package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type TokenAuth struct {
	Usecase *usecase.Usecase
}

// TokenValidation middleware for validating tokens
func (u TokenAuth) TokenValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate token
		sessionToken, err := u.Usecase.IsTokenValid(r)

		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, err.Error()), http.StatusFound)
			return
		}

		// Add the token to the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.AuthToken, sessionToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
