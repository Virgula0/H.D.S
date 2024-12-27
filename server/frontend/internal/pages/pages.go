package pages

import (
	"fmt"
	"html/template"
	"log"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/usecase"
)

type ServiceHandler struct {
	Usecase *usecase.Usecase
}

// NewServiceHandler main microservice entrypoint; initializes templates
func NewServiceHandler(templates *template.Template) (ServiceHandler, error) {

	repo, err := repository.NewRepository()

	if err != nil {
		e := fmt.Errorf("fail NewRepository: %s", err.Error())
		log.Println(e)
		return ServiceHandler{}, e
	}

	uc := usecase.NewUsecase(repo, templates)
	return ServiceHandler{
		Usecase: uc,
	}, nil
}
