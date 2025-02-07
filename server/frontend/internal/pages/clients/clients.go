package clients

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/response"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"net/http"
	"net/url"
)

type Page struct {
	Usecase *usecase.Usecase
}

type ClientTemplate struct {
	Page int `query:"page"`
}

// ListClients renders clients installed by users
func (u Page) ListClients(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	errorMessage := r.URL.Query().Get("error")

	var request ClientTemplate
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidateQueryParameters(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.ClientPage), http.StatusFound)
		return
	}

	var page = 1

	if request.Page != 0 {
		page = request.Page
	}

	clients, err := u.Usecase.GetUserClients(token.(string), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	postsPerPage := 5
	totalPages := (clients.Length + postsPerPage - 1) / postsPerPage

	// Get Clients installed by user

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.ClientView, map[string]any{
		"Clients":     clients.Clients,
		"Certs":       clients.Certs,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Error":       errorMessage,
	})
}

type DeleteClientRequest struct {
	UUID string `form:"uuid" validate:"required"`
}

// DeleteClient Accept post request for deleting a cliemt
func (u Page) DeleteClient(w http.ResponseWriter, r *http.Request) {
	var request DeleteClientRequest
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidatePOSTFormRequest(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.ClientPage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	result, err := u.Usecase.DeleteClientRequest(token.(string), &entities.DeleteClientRequest{
		ClientUUID: request.UUID,
	})

	if err != nil && result != nil && !result.Status {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.ClientPage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.ClientPage), http.StatusFound)
}
