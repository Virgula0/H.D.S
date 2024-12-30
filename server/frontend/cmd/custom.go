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
	cr.PathPrefix(path).Handler(http.StripPrefix(path, customFileServer(http.FS(fsys))))
}

// customFileServer returns a handler that serves HTTP requests with the contents of fsys.
// If a directory is requested, it returns a 404 Not Found error.
func customFileServer(fsys http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := fsys.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if stat.IsDir() {
			// If the request is for a directory, return 404 Not Found
			http.NotFound(w, r)
		} else {
			// Otherwise, serve the file
			http.FileServer(fsys).ServeHTTP(w, r)
		}
	})
}
