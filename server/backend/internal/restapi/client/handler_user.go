package client

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/entities"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type ReturnClientsInstalledByUserResponse struct {
	Length  int                `json:"length"`
	Clients []*entities.Client `json:"clients"`
}

func (u Handler) ReturnClientsInstalledByUser(w http.ResponseWriter, r *http.Request) {
	c := response.ResponseInitializer{ResponseWriter: w}

	userID, err := u.Usecase.GetUserIDFromToken(r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	clientsInstalled, len, err := u.Usecase.GetClientsInstalledByUser(userID.String(), 1) // offset must start from 1

	if len <= 0 {
		c.JSON(http.StatusNotFound, response.UniformResponse{
			StatusCode: http.StatusNotFound,
			Details:    errors.ErrElementNotFound.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.UniformResponse{
			StatusCode: http.StatusInternalServerError,
			Details:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ReturnClientsInstalledByUserResponse{
		Length:  len,
		Clients: clientsInstalled,
	})
}
