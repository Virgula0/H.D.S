package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/frontend/static"
	"github.com/Virgula0/progetto-dp/server/frontend/views"
)

var ServerHost = os.Getenv("FRONTEND_HOST")
var ServerPort = os.Getenv("FRONTEND_PORT")

func runService(router *CustomRouter, templates *template.Template) error {

	ms, err := pages.NewServiceHandler(templates)
	if err != nil {
		e := fmt.Errorf("fail handlers.Handlers: %s", err.Error())
		return e
	}

	// run microservices
	ms.InitRoutes(router.Router)

	return nil

}

var templateMapFunctions = template.FuncMap{
	"add": usecase.AddForTemplate,
	"sub": usecase.SubForTemplate,
	"seq": usecase.SeqForTemplate,
	"lt":  usecase.LtForTemplate,
	"eq":  usecase.EqualForTemplate,
}

func createServer(handler http.Handler, host, port string) *http.Server {
	s := &http.Server{
		Addr:              host + ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	s.SetKeepAlivesEnabled(false)
	return s
}

func RunFrontEnd() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := NewCustomRouter()

	// Parse templates
	templ := template.Must(template.New("").Funcs(templateMapFunctions).ParseFS(views.ViewsFS, "*.html"))

	// Serve Static Files
	stylesSub, _ := fs.Sub(static.StylesFS, "styles")
	scriptsSub, _ := fs.Sub(static.ScriptsFS, "scripts")
	imageSub, _ := fs.Sub(static.ImageFS, "images")
	router.StaticFS("/styles", stylesSub)
	router.StaticFS("/scripts", scriptsSub)
	router.StaticFS("/images", imageSub)

	err := runService(router, templ)

	if err != nil {
		panic(err)
	}

	srv := createServer(router.Router, ServerHost, ServerPort)

	log.Printf("FE ready on %s:%s\n", ServerHost, ServerPort)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}
