package handshakes

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/response"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"io"
	"mime/multipart"
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
	Page int `query:"page"`
}

// TemplateHandshake renders the login page template with an error message (if any)
func (u Page) TemplateHandshake(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	errorMessage := r.URL.Query().Get("error")
	successMessage := r.URL.Query().Get("success")

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

	var page = 1

	if request.Page != 0 {
		page = request.Page
	}

	handshakes, err := u.Usecase.GetUserHandshakes(token.(string), page)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Iterate until data available, we need all clients installed by user
	var clientPage = 1
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
		availableClients = append(availableClients, fmt.Sprintf("%s:%s", client.Name, client.ClientUUID))
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
		"Success":          successMessage,
		"InstalledClients": strings.Join(availableClients, ";"),
	})
}

type UpdateTaskRequest struct {
	AssignedClientUUID string `form:"clientUUID" validate:"required"`
	HandshakeUUID      string `form:"uuid" validate:"required"`
	AttackMode         string `form:"attackMode" validate:"required"`
	HashMode           string `form:"hashMode" validate:"required"`
	Wordlist           string `form:"wordlist"`
	OtherOptions       string `form:"otherOptions"`
}

// UpdateTask Accept post request for updating a task
func (u Page) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var request UpdateTaskRequest
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidatePOSTFormRequest(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	// do checks and then submit
	otherOptions := ""
	if request.OtherOptions != "" {
		otherOptions = " " + request.OtherOptions
	}

	wordlist := ""
	if request.Wordlist != "" {
		wordlist = " " + request.Wordlist
	}

	command := fmt.Sprintf("-a %s -m %s --potfile-disable --logfile-disable %s%s%s", request.AttackMode, request.HashMode, constants.FileToCrackString, wordlist, otherOptions)
	formatted := &entities.UpdateHandshakeTaskViaAPIRequest{
		HandshakeUUID:      request.HandshakeUUID,
		AssignedClientUUID: request.AssignedClientUUID,
		HashcatOptions:     command,
	}
	crackingRequest, err := u.Usecase.SendCrackingRequest(token.(string), formatted)

	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	if !crackingRequest.Success {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, crackingRequest.Reason), http.StatusFound)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1&success=%s", constants.HandshakePage, url.QueryEscape(fmt.Sprintf("%s updated", crackingRequest.Handshake.UUID))), http.StatusFound)
}

type DeleteHandshakeRequest struct {
	UUID string `form:"uuid" validate:"required"`
}

func (u Page) DeleteHandshake(w http.ResponseWriter, r *http.Request) {
	var request DeleteHandshakeRequest
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidatePOSTFormRequest(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	result, err := u.Usecase.DeleteHandshakeRequest(token.(string), &entities.DeleteHandshakesRequest{
		HandshakeUUID: request.UUID,
	})

	if err != nil && result != nil && !result.Status {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1&success=%s", constants.HandshakePage, url.QueryEscape(fmt.Sprintf("%s has deleted", request.UUID))), http.StatusFound)
}

type CreateHandshakeRequest struct {
	File multipart.File `form:"file" validate:"required"`
}

func (u Page) CreateHandshake(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(constants.MaxUploadSize); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("failed to parse multipart form data")), http.StatusFound)
		return
	}

	// Retrieve the file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("file is required")), http.StatusFound)
		return
	}
	defer file.Close()

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("failed to read file")), http.StatusFound)
		return
	}

	hnd := &entities.CreateHandshakeRequest{
		HandshakePCAP: fileBytes,
	}

	id, err := u.Usecase.CreateHandshake(token.(string), hnd)

	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1&success=%s", constants.HandshakePage, url.QueryEscape(fmt.Sprintf("hadnshake %s created", id.HandshakeID))), http.StatusFound)
}
