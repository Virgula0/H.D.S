package middlewares

import (
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
)

// CheckCookieExistence middleware checks for the existence and validity of a session cookie
//
// This middleware just checks to redirect if a valid session_token cookie is already present in the request
// Called by login and register templates avoiding rendering
func (u TokenAuth) CheckCookieExistence(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the session token cookie exists
		_, err := r.Cookie(constants.SessionTokenName)

		if err == nil {
			// If the cookie exists, validate the token
			_, errValidation := u.Usecase.IsTokenValid(r)

			if errValidation == nil {
				// If token is valid, redirect to posts page
				http.Redirect(w, r, constants.HandshakePage, http.StatusFound)
				return
			}

			// If token is invalid, delete the cookie
			http.SetCookie(w, &http.Cookie{
				Name:     constants.SessionTokenName,
				Value:    "",
				Expires:  time.Unix(0, 0),
				Path:     "/",
				HttpOnly: true,
			})
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}
