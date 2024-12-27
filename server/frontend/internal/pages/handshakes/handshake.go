package handshakes

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/response"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"net/http"
	"net/url"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type Page struct {
	Usecase *usecase.Usecase
}

type HandshakeTemplate struct {
	Page uint `query:"page"`
}

// ListHandshakes renders the login page template with an error message (if any)
func (u Page) ListHandshakes(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	var request HandshakeTemplate
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidateQueryParameters(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.HandshakePage), http.StatusFound)
		return
	}

	page := 1

	if request.Page != 0 {
		page = int(request.Page)
	}

	handshakes, err := u.Usecase.GetUserHandshakes(token.(string), page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	postsPerPage := 5
	totalPages := (handshakes.Length + postsPerPage - 1) / postsPerPage

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.HandshakeView, map[string]any{
		"Handshakes":  handshakes.Handshakes,
		"CurrentPage": page,
		"TotalPages":  totalPages,
	})
}
