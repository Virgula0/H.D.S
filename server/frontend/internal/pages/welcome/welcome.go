package welcome

import (
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
	"net/http"
)

type Page struct {
	Usecase *usecase.Usecase
}

func (u Page) WelcomeTemplate(w http.ResponseWriter, r *http.Request) {
	u.Usecase.RenderTemplate(w, constants.WelcomeView, nil)
}
