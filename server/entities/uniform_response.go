package entities

// UniformResponse is used to provide a uniform correct message structure from API
type UniformResponse struct {
	StatusCode int    `json:"status_code"`
	Details    string `json:"details"`
} // @name UniformResponse
