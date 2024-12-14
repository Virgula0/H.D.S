package response

import (
	"encoding/json"
	"html"
	"net/http"

	"log"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
)

// UniformResponse is used to provide a uniform correct message structure from API
type UniformResponse struct {
	StatusCode int    `json:"status_code"`
	Details    string `json:"details"`
} // @name UniformResponse

type ResponseInitializer struct {
	http.ResponseWriter
}

func (w *ResponseInitializer) JSON(statusCode int, toMarshal any) {
	w.Header().Set("Content-Type", constants.JSON_CONTENT_TYPE)
	w.WriteHeader(statusCode)

	// Check Responses structure types and sanitize.
	switch v := toMarshal.(type) {
	case UniformResponse:
		v.Details = html.EscapeString(v.Details)
		toMarshal = v // v is a shallow copy of toMarshal need re-assignment after changes
	}

	marshaled, err := json.Marshal(toMarshal)

	if err != nil {
		log.Printf("[ERROR] While marshaling -> %s", err.Error())
	}

	w.Write(marshaled)
}
