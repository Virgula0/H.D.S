package handshake

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type Handler struct {
	Usecase *usecase.Usecase
}

type GetHandshakeResponse struct {
	Length     int `json:"length"`
	Handshakes []*entities.Handshake
}

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

	handshakes, counted, err := u.Usecase.GetHandshakes(userID.String(), 1) // TODO: handle offset from request

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

	c.JSON(http.StatusOK, GetHandshakeResponse{
		Length:     counted,
		Handshakes: handshakes,
	})
}
