package response

import (
	"encoding/json"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	log "github.com/sirupsen/logrus"
	"html"
	"net/http"
)

type Initializer struct {
	http.ResponseWriter
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (w *Initializer) JSON(statusCode int, toMarshal any) {
	w.Header().Set("Content-Type", constants.JSONContentType)
	w.WriteHeader(statusCode)

	// Check Responses structure types and sanitize.
	switch v := toMarshal.(type) {
	case entities.UniformResponse:
		v.Details = html.EscapeString(v.Details)
		toMarshal = v // v is a shallow copy of toMarshal need re-assignment after changes
	case []*entities.Client:
		for _, c := range v {
			c.Name = html.EscapeString(c.Name)
			c.LatestIP = html.EscapeString(c.LatestIP)
		}
		toMarshal = v
	}

	marshaled, err := json.Marshal(toMarshal)

	if err != nil {
		log.Printf("[ERROR] While marshaling -> %s", err.Error())
	}

	_, err = w.Write(marshaled)

	if err != nil {
		log.Panic(err.Error())
		return
	}
}
