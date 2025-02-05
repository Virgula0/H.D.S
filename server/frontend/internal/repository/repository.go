package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
)

/*
repository.go contains simple methods for performing requests to the backend
*/

type Repository struct {
	client *http.Client
}

// CustomTransport wraps around an existing http.RoundTripper
type CustomTransport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and sets default headers
func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Set default headers here
	req.Header.Set("Content-Type", "application/json")

	// Use the embedded RoundTripper to execute the actual request
	return c.Transport.RoundTrip(req)
}

func NewRepository() (*Repository, error) {
	return &Repository{
		client: &http.Client{
			Timeout: time.Second * constants.TimeOut,
			Transport: &CustomTransport{
				Transport: http.DefaultTransport,
			},
		},
	}, nil
}

func (repo *Repository) GenericHTTPRequest(baseURL, method, endpoint string, headers map[string]string, requestData []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, fmt.Sprintf("%s%s", baseURL, endpoint), bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (repo *Repository) GenericHTTPRequestToBackend(method, endpoint string, headers map[string]string, requestData []byte) ([]byte, error) {
	return repo.GenericHTTPRequest(constants.BackendBaseURL, method, endpoint, headers, requestData)
}

func (repo *Repository) uniformResponseRefactored(requestData any, endpoint, method string, headers map[string]string) (*entities.UniformResponse, error) {

	// Marshal the data into JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	responseBody, err := repo.GenericHTTPRequestToBackend(method, endpoint, headers, jsonData)

	if err != nil {
		return nil, err
	}

	var backendResponse entities.UniformResponse

	err = json.Unmarshal(responseBody, &backendResponse)
	if err != nil {
		return nil, err
	}

	return &backendResponse, nil
}

func (repo *Repository) PerformLogin(username, password string) (*entities.UniformResponse, error) {
	requestData := map[string]string{
		"username": username,
		"password": password,
	}

	return repo.uniformResponseRefactored(requestData, constants.BackendAuthEndpoint, http.MethodPost, nil)
}

func (repo *Repository) PerformLogout(token string) (*entities.UniformResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	return repo.uniformResponseRefactored(nil, constants.BackendLogoutEndpoint, http.MethodGet, headers)
}

func (repo *Repository) PerformRegister(username, password, confirmation string) (*entities.UniformResponse, error) {
	requestData := map[string]string{
		"username":     username,
		"password":     password,
		"confirmation": confirmation,
	}

	return repo.uniformResponseRefactored(requestData, constants.BackendRegisterEndpoint, http.MethodPost, nil)
}

func (repo *Repository) GetUserHandshakes(token string, page int) (*entities.GetHandshakeResponse, error) {
	var handshakes *entities.GetHandshakeResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodGet, fmt.Sprintf("%s?page=%s", constants.BackendGetHandshakes, strconv.Itoa(page)), headers, nil)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &handshakes); err != nil {
		return nil, err
	}

	return handshakes, nil
}

func (repo *Repository) GetUserClients(token string, page int) (*entities.ReturnClientsInstalledResponse, error) {
	var clients *entities.ReturnClientsInstalledResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodGet, fmt.Sprintf("%s?page=%s", constants.BackendGetClients, strconv.Itoa(page)), headers, nil)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (repo *Repository) GetUserDevices(token string, page int) (*entities.ReturnRaspberryPiDevicesResponse, error) {
	var clients *entities.ReturnRaspberryPiDevicesResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodGet, fmt.Sprintf("%s?page=%s", constants.BackendGetRaspberryPi, strconv.Itoa(page)), headers, nil)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (repo *Repository) SendCrackingRequest(token string, request *entities.UpdateHandshakeTaskViaAPIRequest) (*entities.UpdateHandshakeTaskViaAPIResponse, error) {
	var update *entities.UpdateHandshakeTaskViaAPIResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodPost, constants.BackendUpdateClientTask, headers, jsonData)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &update); err != nil {
		return nil, err
	}

	return update, nil
}

func (repo *Repository) DeleteClient(token string, request *entities.DeleteClientRequest) (*entities.DeleteClientResponse, error) {
	var dd *entities.DeleteClientResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodDelete, constants.BackendDeleteClient, headers, jsonData)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &dd); err != nil {
		return nil, err
	}

	return dd, nil
}

func (repo *Repository) DeleteRaspberryPI(token string, request *entities.DeleteRaspberryPIRequest) (*entities.DeleteRaspberryPIResponse, error) {
	var dd *entities.DeleteRaspberryPIResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodDelete, constants.BackendDeleteRaspberryPI, headers, jsonData)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &dd); err != nil {
		return nil, err
	}

	return dd, nil
}

func (repo *Repository) DeleteHandshake(token string, request *entities.DeleteHandshakesRequest) (*entities.DeleteHandshakesResponse, error) {
	var dd *entities.DeleteHandshakesResponse

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodDelete, constants.BackendHandshake, headers, jsonData)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &dd); err != nil {
		return nil, err
	}

	return dd, nil
}

func (repo *Repository) CreateHandshake(token string, request *entities.CreateHandshakeRequest) (*entities.CreateHandshakeResponse, error) {

	var dd *entities.CreateHandshakeResponse
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodPut, constants.BackendHandshake, headers, jsonData)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseBytes, &dd); err != nil {
		return nil, err
	}

	return dd, nil
}
