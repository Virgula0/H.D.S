package authenticate

import (
	"net/http"

	rr "github.com/Virgula0/progetto-dp/server/backend/internal/response"
)

func (u Handler) ChekTokenValidity(w http.ResponseWriter, r *http.Request) {

	c := rr.ResponseInitializer{ResponseWriter: w}

	c.JSON(http.StatusOK, rr.UniformResponse{
		StatusCode: http.StatusOK,
		Details:    "valid",
	})
}
