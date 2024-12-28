package login

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	rr "github.com/Virgula0/progetto-dp/server/frontend/internal/response"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"net/http"
	"net/url"
	"time"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type Page struct {
	Usecase *usecase.Usecase
}

// LoginTemplate renders the login page template with an error message (if any)
func (u Page) LoginTemplate(w http.ResponseWriter, r *http.Request) {
	errorMessage := r.URL.Query().Get("error")

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.LoginView, map[string]interface{}{
		"Error": errorMessage,
	})
}

type PerformLogin struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}

func (u Page) PerformLogin(w http.ResponseWriter, r *http.Request) {
	c := rr.Initializer{ResponseWriter: w}

	// Validate form input
	var loginRequest PerformLogin
	if err := utils.ValidatePOSTFormRequest(&loginRequest, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	// Perform login logic
	loginResponse, err := u.Usecase.PerformLogin(loginRequest.Username, loginRequest.Password)
	if err != nil {
		errorQuery := url.QueryEscape(err.Error())
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, errorQuery), http.StatusFound)
		return
	}

	if loginResponse.StatusCode != http.StatusOK {
		errorQuery := url.QueryEscape(errors.ErrInvalidCredentials.Error()) // Ensures the string is URL-safe
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, errorQuery), http.StatusFound)
		return
	}

	// Set the session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionTokenName,
		Value:    loginResponse.Details,
		Expires:  time.Now().Add(3 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect to posts
	http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.HandshakePage), http.StatusFound)
}