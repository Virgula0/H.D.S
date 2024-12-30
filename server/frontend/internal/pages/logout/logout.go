package pages

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type Page struct {
	Usecase *usecase.Usecase
}

func (u Page) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(constants.AuthToken)

	// Check if the token exists
	if token == nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, url.QueryEscape(errors.ErrNotAuthenticated.Error())), http.StatusFound)
		return
	}

	logout, err := u.Usecase.PerformLogout(token.(string))

	if err != nil {
		errorQuery := url.QueryEscape(err.Error()) // Ensures the string is URL-safe
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, errorQuery), http.StatusFound)
		return
	}

	if logout.StatusCode != http.StatusOK {
		errorQuery := url.QueryEscape(errors.ErrLogout.Error()) // Ensures the string is URL-safe
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", constants.Login, errorQuery), http.StatusFound)
		return
	}

	cookie := http.Cookie{
		Name:     constants.SessionTokenName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	http.Redirect(w, r, constants.Login, http.StatusFound)
}
