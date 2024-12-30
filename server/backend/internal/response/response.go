package response

import (
	"encoding/json"
	"html"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type Initializer struct {
	http.ResponseWriter
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
		log.Errorf("[ERROR] While marshaling -> %s", err.Error())
	}

	_, err = w.Write(marshaled)

	if err != nil {
		log.Panic(err.Error())
		return
	}
}
