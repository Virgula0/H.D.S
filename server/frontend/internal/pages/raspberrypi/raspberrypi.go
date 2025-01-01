package raspberrypi

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

// ListRaspberryPI list raspberry pi instealled by the user
func (u Page) ListRaspberryPI(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.RaspberryPIPage), http.StatusFound)
		return
	}

	var page = 1

	if request.Page != 0 {
		page = request.Page
	}

	devices, err := u.Usecase.GetUserDevices(token.(string), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	postsPerPage := 5
	totalPages := (devices.Length + postsPerPage - 1) / postsPerPage

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.DeviceView, map[string]any{
		"RaspberryPis": devices.Devices,
		"CurrentPage":  page,
		"TotalPages":   totalPages,
		"Error":        errorMessage,
	})
}

type DeleteRaspberryPIRequest struct {
	UUID string `form:"uuid" validate:"required"`
}

func (u Page) DeleteRaspberryPI(w http.ResponseWriter, r *http.Request) {
	var request DeleteRaspberryPIRequest
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.Login, url.QueryEscape(customErrors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	if err := utils.ValidatePOSTFormRequest(&request, r); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.RaspberryPIPage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	result, err := u.Usecase.DeleteRaspberryPIRequest(token.(string), &entities.DeleteRaspberryPIRequest{
		RaspberryPIUUID: request.UUID,
	})

	if err != nil && result != nil && !result.Status {
		http.Redirect(w, r, fmt.Sprintf("%s?page=1&error=%s", constants.RaspberryPIPage, url.QueryEscape(err.Error())), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?page=1", constants.RaspberryPIPage), http.StatusFound)
}
