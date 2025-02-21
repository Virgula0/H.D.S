package handshake

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
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

type GetHandshakesRequest struct {
	Page uint `query:"page" validate:"required,min=1"`
}

// GetHandshakes handles logic for getting user's handshakes
func (u Handler) GetHandshakes(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request GetHandshakesRequest

	if err = utils.ValidateQueryParameters(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	handshakes, counted, err := u.Usecase.GetHandshakes(userID.String(), request.Page)

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

	c.JSON(http.StatusOK, entities.GetHandshakeResponse{
		Length:     counted,
		Handshakes: handshakes,
	})
}

// UpdateClientTask handles logic for updating handshake status
func (u Handler) UpdateClientTask(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.UpdateHandshakeTaskViaAPIRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	task, err := u.Usecase.UpdateClientTaskRest(userID.String(), request.HandshakeUUID, request.AssignedClientUUID, constants.PendingStatus, request.HashcatOptions, "", "")
	if err != nil {
		c.JSON(http.StatusOK, entities.UpdateHandshakeTaskViaAPIResponse{
			Success:   false,
			Reason:    err.Error(),
			Handshake: task,
		})
		return
	}

	c.JSON(http.StatusOK, entities.UpdateHandshakeTaskViaAPIResponse{
		Success:   true,
		Handshake: task,
	})
}

// DeleteHandshake handles logic for deleting an handshake
func (u Handler) DeleteHandshake(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.DeleteHandshakesRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	deleted, err := u.Usecase.DeleteHandshake(userID.String(), request.HandshakeUUID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.DeleteHandshakesResponse{
		Status: deleted,
	})
}

func (u Handler) CreateHandshake(w http.ResponseWriter, r *http.Request) {
	c := response.Initializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	var request entities.CreateHandshakeRequest

	if err = utils.ValidateJSON(&request, r); err != nil {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    err.Error(),
		})
		return
	}

	if len(request.HandshakePCAP) > errors.MaxUploadSize {
		c.JSON(http.StatusBadRequest, entities.UniformResponse{
			StatusCode: http.StatusBadRequest,
			Details:    errors.ErrFileTooBig.Error(),
		})
		return
	}

	handshake, err := u.Usecase.CreateHandshake(userID.String(), "", "", constants.NothingStatus, utils.BytesToBase64String(request.HandshakePCAP))
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entities.CreateHandshakeResponse{
		HandshakeID: handshake,
	})
}
