package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
)

type Repository struct {
	client *http.Client
}

type CustomTransport struct {
	Transport http.RoundTripper
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
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

// Helper function to check for UniformResponse errors
func (repo *Repository) checkUniformError(responseBytes []byte) (int, error) {
	var uniformResponse entities.UniformResponse
	if err := json.Unmarshal(responseBytes, &uniformResponse); err != nil {
		return -1, err
	}
	if uniformResponse.Details != "" {
		return uniformResponse.StatusCode, errors.New(uniformResponse.Details)
	}
	return -1, nil
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

	return io.ReadAll(resp.Body)
}

func (repo *Repository) GenericHTTPRequestToBackend(method, endpoint string, headers map[string]string, requestData []byte) ([]byte, error) {
	return repo.GenericHTTPRequest(constants.BackendBaseURL, method, endpoint, headers, requestData)
}

// Common handler for endpoints returning UniformResponse
func (repo *Repository) uniformResponseRefactored(requestData any, endpoint, method string, headers map[string]string) (*entities.UniformResponse, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	responseBody, err := repo.GenericHTTPRequestToBackend(method, endpoint, headers, jsonData)
	if err != nil {
		return nil, err
	}

	var backendResponse entities.UniformResponse
	if err := json.Unmarshal(responseBody, &backendResponse); err != nil {
		return nil, err
	}

	return &backendResponse, nil
}

// Authentication handlers
func (repo *Repository) PerformLogin(username, password string) (*entities.UniformResponse, error) {
	return repo.uniformResponseRefactored(
		map[string]string{"username": username, "password": password},
		constants.BackendAuthEndpoint,
		http.MethodPost,
		nil,
	)
}

func (repo *Repository) PerformLogout(token string) (*entities.UniformResponse, error) {
	return repo.uniformResponseRefactored(
		nil,
		constants.BackendLogoutEndpoint,
		http.MethodGet,
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)},
	)
}

func (repo *Repository) PerformRegister(username, password, confirmation string) (*entities.UniformResponse, error) {
	return repo.uniformResponseRefactored(
		map[string]string{"username": username, "password": password, "confirmation": confirmation},
		constants.BackendRegisterEndpoint,
		http.MethodPost,
		nil,
	)
}

// Data retrieval handlers with pagination
func (repo *Repository) getPaginatedResource(token, endpoint string, page int, target interface{}) error {
	headers := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)}
	endpointWithPage := fmt.Sprintf("%s?page=%d", endpoint, page)

	responseBytes, err := repo.GenericHTTPRequestToBackend(http.MethodGet, endpointWithPage, headers, nil)
	if err != nil {
		return err
	}

	if statusCode, err := repo.checkUniformError(responseBytes); err != nil && statusCode != http.StatusNotFound {
		return err
	}

	return json.Unmarshal(responseBytes, target)
}

func (repo *Repository) GetUserHandshakes(token string, page int) (*entities.GetHandshakeResponse, error) {
	var response entities.GetHandshakeResponse
	err := repo.getPaginatedResource(token, constants.BackendGetHandshakes, page, &response)
	return &response, err
}

func (repo *Repository) GetUserClients(token string, page int) (*entities.ReturnClientsInstalledResponse, error) {
	var response entities.ReturnClientsInstalledResponse
	err := repo.getPaginatedResource(token, constants.BackendGetClients, page, &response)
	return &response, err
}

func (repo *Repository) GetUserDevices(token string, page int) (*entities.ReturnRaspberryPiDevicesResponse, error) {
	var response entities.ReturnRaspberryPiDevicesResponse
	err := repo.getPaginatedResource(token, constants.BackendGetRaspberryPi, page, &response)
	return &response, err
}

// Common CRUD operation handler
func (repo *Repository) executeAuthorizedRequest(method, endpoint, token string, request, response any) error {
	headers := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	responseBytes, err := repo.GenericHTTPRequestToBackend(method, endpoint, headers, jsonData)
	if err != nil {
		return err
	}

	if statusCode, err := repo.checkUniformError(responseBytes); err != nil && statusCode != http.StatusNotFound {
		return err
	}

	return json.Unmarshal(responseBytes, response)
}

// Cracking operations
func (repo *Repository) SendCrackingRequest(token string, request *entities.UpdateHandshakeTaskViaAPIRequest) (*entities.UpdateHandshakeTaskViaAPIResponse, error) {
	var response entities.UpdateHandshakeTaskViaAPIResponse
	err := repo.executeAuthorizedRequest(http.MethodPost, constants.BackendUpdateClientTask, token, request, &response)
	return &response, err
}

// Deletion operations
func (repo *Repository) DeleteClient(token string, request *entities.DeleteClientRequest) (*entities.DeleteClientResponse, error) {
	var response entities.DeleteClientResponse
	err := repo.executeAuthorizedRequest(http.MethodDelete, constants.BackendDeleteClient, token, request, &response)
	return &response, err
}

func (repo *Repository) DeleteRaspberryPI(token string, request *entities.DeleteRaspberryPIRequest) (*entities.DeleteRaspberryPIResponse, error) {
	var response entities.DeleteRaspberryPIResponse
	err := repo.executeAuthorizedRequest(http.MethodDelete, constants.BackendDeleteRaspberryPI, token, request, &response)
	return &response, err
}

func (repo *Repository) DeleteHandshake(token string, request *entities.DeleteHandshakesRequest) (*entities.DeleteHandshakesResponse, error) {
	var response entities.DeleteHandshakesResponse
	err := repo.executeAuthorizedRequest(http.MethodDelete, constants.BackendHandshake, token, request, &response)
	return &response, err
}

// Creation operations
func (repo *Repository) CreateHandshake(token string, request *entities.CreateHandshakeRequest) (*entities.CreateHandshakeResponse, error) {
	var response entities.CreateHandshakeResponse
	err := repo.executeAuthorizedRequest(http.MethodPut, constants.BackendHandshake, token, request, &response)
	return &response, err
}

func (repo *Repository) UploadWordlist(token string, request *entities.UploadWordlistRequest) (*entities.UploadWordlistResponse, error) {
	var response entities.UploadWordlistResponse
	err := repo.executeAuthorizedRequest(http.MethodPut, constants.UploadWordlistBackend, token, request, &response)
	return &response, err
}

// Update operations
func (repo *Repository) UpdateClientEncryptionStatus(token string, request *entities.UpdateEncryptionClientStatusRequest) (*entities.UpdateEncryptionClientStatusResponse, error) {
	var response entities.UpdateEncryptionClientStatusResponse
	err := repo.executeAuthorizedRequest(http.MethodPost, constants.UpdateClientEncryption, token, request, &response)
	return &response, err
}

func (repo *Repository) UpdateUserPassword(token string, request *entities.UpdateUserPasswordRequest) (*entities.UpdateUserPasswordResponse, error) {
	var response entities.UpdateUserPasswordResponse
	err := repo.executeAuthorizedRequest(http.MethodPost, constants.UpdateUserPassword, token, request, &response)
	return &response, err
}
