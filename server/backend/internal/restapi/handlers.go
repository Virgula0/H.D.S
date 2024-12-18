package restapi

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/entities"
	"log"

	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
)

type ServiceHandler struct {
	Usecase *usecase.Usecase
}

// NewServiceHandler main microservice entrypoint; creates repository, seeds database and exposes usecase
func NewServiceHandler(db *infrastructure.Database) (ServiceHandler, error) {

	repo, err := repository.NewRepository(db)
	if err != nil {
		e := fmt.Errorf("fail NewRepository: %s", err.Error())
		log.Println(e)
		return ServiceHandler{}, e
	}

	if constants.WipeTables != "" {

		// Delete data from DB
		err = db.CleanDB([]string{entities.UserTableName})

		if err != nil {
			return ServiceHandler{}, err
		}

		// tables have been wiped, needs a seed
		seedArray := []error{
			seed.LoadUsers(repo),
		}

		for _, err := range seedArray {
			if err != nil {
				e := fmt.Errorf("fail seed.Load: %s", err.Error())
				log.Println(e)
				panic(err)
			}
		}
	}

	uc := usecase.NewUsecase(repo)
	return ServiceHandler{
		Usecase: uc,
	}, nil
}
