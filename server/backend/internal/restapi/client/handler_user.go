package client

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type Handler struct {
	Usecase *usecase.Usecase
}
type ReturnClientDevicesRequest struct {
	Page uint `query:"page" validate:"required,min=1"`
}

// ReturnClientsInstalled handles logic for returning installed clients
func (u Handler) ReturnClientsInstalled(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request ReturnClientDevicesRequest

	if err = utils.ValidateQueryParameters(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	clientsInstalled, counted, err := u.Usecase.GetClientsInstalledByUserID(userID.String(), request.Page)

	if counted == 0 {
		c.JSON(http.StatusNotFound, entities.UniformResponse{
			StatusCode: http.StatusNotFound,
			Details:    errors.ErrElementNotFound.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	// get all certs by userID
	certsInstalled, _, err := u.Usecase.GetClientCertsByUserID(userID.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.ReturnClientsInstalledResponse{
		Length:  counted,
		Clients: clientsInstalled,
		Certs:   certsInstalled,
	})
}

// DeleteClient handles logic for deleting a existing client
func (u Handler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.DeleteClientRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	deleted, err := u.Usecase.DeleteClient(userID.String(), request.ClientUUID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.DeleteClientResponse{
		Status: deleted,
	})
}

// UpdateEncryptionClientStatus update encryption client status
func (u Handler) UpdateEncryptionClientStatus(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.UpdateEncryptionClientStatusRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	if request.Status == nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    "status must be present",
		})
		return
	}

	err = u.Usecase.UpdateEncryptionClientStatus(request.ClientUUID, userID.String(), *request.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, entities.UniformResponse{
			StatusCode: http.StatusNotFound,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.UpdateEncryptionClientStatusResponse{
		Status: *request.Status,
	})
}
