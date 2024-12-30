package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/utils"
	"html/template"
	"net/http"
)

const genericErrorMessage = "Token+not+valid+or+expired"

type Usecase struct {
	repo      *repository.Repository
	templates *template.Template
}

func NewUsecase(repo *repository.Repository, templates *template.Template) *Usecase {
	return &Usecase{
		repo:      repo,
		templates: templates,
	}
}

// RenderTemplate renders an HTML template
func (uc Usecase) RenderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", constants.HTMLContentType)
	err := uc.templates.ExecuteTemplate(w, name, data)

	if err != nil {
		http.Error(w, fmt.Sprintf("template error %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// TEMPLATING FUNCTIONS

func EqualStringForTemplate(a, b string) bool {
	return a == b
}

func EqualForTemplate(a, b int) bool {
	return a == b
}

func AddForTemplate(a, b int) int {
	return a + b
}

func SubForTemplate(a, b int) int {
	return a - b
}

func SeqForTemplate(start, end int) []int {
	var sequence []int
	for i := start; i <= end; i++ {
		sequence = append(sequence, i)
	}
	return sequence
}

func LtForTemplate(x, y int) bool {
	return x < y
}

// USECASE MAIN FUNCTIONS

func (uc Usecase) IsTokenValid(r *http.Request) (string, error) {
	sessionToken, err := r.Cookie(constants.SessionTokenName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", fmt.Errorf("cookie %s not found", constants.SessionTokenName)
		}
		return "", err
	}

	token := sessionToken.Value

	if !utils.IsJWT(token) {
		return "", fmt.Errorf("cookie not set ot not a valid jwt")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	responseBody, err := uc.GenericHTTPRequest(http.MethodGet, constants.BackendVerifyEndpoint, headers, nil)

	if err != nil {
		return "", err
	}

	var response entities.UniformResponse

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", genericErrorMessage)
	}

	return token, nil
}

func (uc Usecase) GenericHTTPRequest(method, url string, headers map[string]string, requestData []byte) ([]byte, error) {
	return uc.repo.GenericHTTPRequestToBackend(method, url, headers, requestData)
}

func (uc Usecase) PerformLogin(username, password string) (*entities.UniformResponse, error) {
	return uc.repo.PerformLogin(username, password)
}

func (uc Usecase) PerformLogout(token string) (*entities.UniformResponse, error) {
	return uc.repo.PerformLogout(token)
}

func (uc Usecase) PerformRegistration(username, password, confirmation string) (*entities.UniformResponse, error) {
	return uc.repo.PerformRegister(username, password, confirmation)
}

func (uc Usecase) GetUserHandshakes(token string, page int) (*entities.GetHandshakeResponse, error) {
	return uc.repo.GetUserHandshakes(token, page)
}

func (uc Usecase) GetUserClients(token string, page int) (*entities.ReturnClientsInstalledResponse, error) {
	return uc.repo.GetUserClients(token, page)
}
func (uc Usecase) GetUserDevices(token string, page int) (*entities.ReturnRaspberryPiDevicesResponse, error) {
	return uc.repo.GetUserDevices(token, page)
}

func (uc Usecase) SendCrackingRequest(token string, request *entities.UpdateHandshakeTaskViaAPIRequest) (*entities.UpdateHandshakeTaskViaAPIResponse, error) {
	return uc.repo.SendCrackingRequest(token, request)
}
