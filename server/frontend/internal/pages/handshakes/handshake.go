package handshakes

import (
	"net/http"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type Page struct {
	Usecase *usecase.Usecase
}

// HandshakeTemplate renders the login page template with an error message (if any)
func (u Page) HandshakeTemplate(w http.ResponseWriter, r *http.Request) {
	errorMessage := r.URL.Query().Get("error")

	// RenderTemplate the login template
	u.Usecase.RenderTemplate(w, constants.HandshakeView, map[string]interface{}{
		"Error": errorMessage,
	})
}
