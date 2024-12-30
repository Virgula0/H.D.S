package middlewares

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

// LogginMiddlware just logs all incoming requests
func LogginMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Printf("[FRONTEND] - [%s] on %s coming from %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
