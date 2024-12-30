package register

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"net/http"
	"net/url"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type Page struct {
	Usecase *usecase.Usecase
}

func (u Page) Register(w http.ResponseWriter, r *http.Request) {
	errorMessage := r.URL.Query().Get("error")

	u.Usecase.RenderTemplate(w, constants.RegisterView, map[string]interface{}{
		"Error": errorMessage,
	})
}

type PerformRegistrationRequest struct {
	Username     string `form:"username" binding:"required"`
	Password     string `form:"password" binding:"required"`
	Confirmation string `form:"confirmation" binding:"required"`
}

func (u Page) PerformRegistration(w http.ResponseWriter, r *http.Request) {
	var request PerformRegistrationRequest

	// Validate form input
	if err := utils.ValidatePOSTFormRequest(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Register, err.Error()), http.StatusFound)
		return
	}

	registerResponse, err := u.Usecase.PerformRegistration(request.Username, request.Password, request.Confirmation)

	if err != nil {
		errorQuery := url.QueryEscape(err.Error()) // Ensures the string is URL-safe
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Register, errorQuery), http.StatusFound)
		return
	}

	if registerResponse.StatusCode != http.StatusOK {
		errorQuery := url.QueryEscape(registerResponse.Details) // Ensures the string is URL-safe
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Register, errorQuery), http.StatusFound)
		return
	}

	// Redirect to login
	http.Redirect(w, r, constants.Login, http.StatusFound)
}
