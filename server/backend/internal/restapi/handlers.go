package restapi

import (
	"fmt"
	"log"

	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/seed"
	"github.com/Virgula0/progetto-dp/server/backend/internal/usecase"
	"github.com/Virgula0/progetto-dp/server/entities"
)

type ServiceHandler struct {
	Usecase *usecase.Usecase
}

// NewServiceHandler main microservice entrypoint; creates repository, seeds database and exposes usecase
func NewServiceHandler(db *infrastructure.Database, reset bool) (ServiceHandler, error) {

	repo, err := repository.NewRepository(db, reset)
	if err != nil {
		e := fmt.Errorf("fail NewRepository: %s", err.Error())
		log.Println(e)
		return ServiceHandler{}, e
	}

	if reset {

		// wipe tables first, if requested
		cleanTables := []string{
			fmt.Sprintf("DELETE FROM %s", entities.UserTableName),
		}

		for _, query := range cleanTables {
			_, err := db.Exec(query)
			if err != nil {
				return ServiceHandler{}, fmt.Errorf("unable to exec delete query %s ERROR: %v", query, err)
			}
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
