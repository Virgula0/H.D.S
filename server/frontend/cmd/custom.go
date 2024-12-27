package cmd

import (
	"github.com/gorilla/mux"
	"io/fs"
	"net/http"
)

// CustomRouter wraps Gorilla Mux and adds Gin-like functionality
type CustomRouter struct {
	*mux.Router
}

// NewCustomRouter initializes a new CustomRouter
func NewCustomRouter() *CustomRouter {
	return &CustomRouter{
		Router: mux.NewRouter(),
	}
}

// StaticFS serves static files from a filesystem
func (cr *CustomRouter) StaticFS(path string, fsys fs.FS) {
	cr.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(http.FS(fsys))))
}
