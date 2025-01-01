package middlewares

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

// LoggingMiddleware just logs all incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Printf("[REST-API-REQUEST] - [%s] on %s coming from %s", r.Method, r.RequestURI, r.RemoteAddr)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
