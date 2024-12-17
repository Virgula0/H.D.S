package testsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/Virgula0/progetto-dp/server/entities"
	"io"
	"net/http"
	"net/url"
	"time"
)

var APILOGIN = fmt.Sprintf("http://%s:%s/v1/auth", "localhost", "4747")

// HTTPRequest performs an HTTP request with the specified method, URL, headers, query parameters, and body.
// It returns the response body as a string and an error if any.
func HTTPRequest(method, urlStr string, headers map[string]string, queryParams url.Values, body []byte, timeout time.Duration) (string, error) {
	// Add query parameters to the URL
	if queryParams != nil {
		urlStr = fmt.Sprintf("%s?%s", urlStr, queryParams.Encode())
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, urlStr, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type if not provided
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code indicates an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body as a string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}

func AuthAPI(auth authenticate.AuthRequest) (string, error) {
	marshaled, err := json.Marshal(&auth)

	if err != nil {
		return "", err
	}

	response, err := HTTPRequest("POST", APILOGIN, map[string]string{}, url.Values{}, marshaled, 10*time.Second)

	if err != nil {
		return "", err
	}

	var authResponse entities.UniformResponse

	err = json.Unmarshal([]byte(response), &authResponse)

	if err != nil {
		return "", err
	}

	return authResponse.Details, nil
}
