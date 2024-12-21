package authenticate

import (
	"github.com/Virgula0/progetto-dp/server/entities"
	"net/http"

	rr "github.com/Virgula0/progetto-dp/server/backend/internal/response"
)

func (u Handler) CheckTokenValidity(w http.ResponseWriter, r *http.Request) {

	c := rr.Initializer{ResponseWriter: w}

	c.JSON(http.StatusOK, entities.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "valid",
	})
}
