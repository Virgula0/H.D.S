package handshakes

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/response"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"net/http"
	"net/url"
	"strings"

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

	errorMessage := r.URL.Query().Get("error")

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

	// Iterate until data available, we need all clients installed by user
	clientPage := 1
	available := 1
	var clients = make([]*entities.Client, 0)

	for available > 0 {
		cc, err := u.Usecase.GetUserClients(token.(string), clientPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		clientPage++
		available = len(cc.Clients)
		clients = append(clients, cc.Clients...)
	}

	availableClients := make([]string, 0)
	for _, client := range clients {
		availableClients = append(availableClients, client.ClientUUID)
	}

	postsPerPage := 5
	totalPages := (handshakes.Length + postsPerPage - 1) / postsPerPage

	// Get Clients installed by user

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.HandshakeView, map[string]any{
		"Handshakes":       handshakes.Handshakes,
		"CurrentPage":      page,
		"TotalPages":       totalPages,
		"Error":            errorMessage,
		"InstalledClients": strings.Join(availableClients, ";"),
	})
}
