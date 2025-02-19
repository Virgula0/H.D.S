package wordlist

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"io"
	"net/http"
	"net/url"
)

type Page struct {
	Usecase *usecase.Usecase
}

type UploadWordlistRequest struct {
	FileName   string `form:"fileName" validate:"required"`
	ClientUUID string `form:"clientUUID" validate:"required"`
}

// UploadWordlist Accept post request for uploading a wordlist
func (u Page) UploadWordlist(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	var request UploadWordlistRequest

	// Parse multipart form data
	if err := r.ParseMultipartForm(constants.MaxUploadSize); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("failed to parse multipart form data")), http.StatusFound)
		return
	}

	// Retrieve the file
	file, _, err := r.FormFile("fileBytes")
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("file is required")), http.StatusFound)
		return
	}
	defer file.Close()

	// Parse remaining normal post
	if err := utils.ValidatePOSTFieldsFromMultipartFormData(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("failed to parse post data from multipart")), http.StatusFound)
		return
	}

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape("failed to read file")), http.StatusFound)
		return
	}

	ww := &entities.UploadWordlistRequest{
		FileBytes:  fileBytes,
		FileName:   request.FileName,
		ClientUUID: request.ClientUUID,
	}

	response, err := u.Usecase.UploadWordlist(token.(string), ww)

	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.HandshakePage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1&success=%s", constants.HandshakePage, url.QueryEscape(fmt.Sprintf("wordlist %s uplaoded. The client will download the wordlist as soon as possible", response.WordlistID))), http.StatusFound)
}
