package authenticate

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/backend/internal/response"
)

func (u Handler) ChekTokenValidity(w http.ResponseWriter, r *http.Request) {

	c := response.ResponseInitializer{ResponseWriter: w}

	c.JSON(http.StatusOK, response.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "valid",
	})
}
