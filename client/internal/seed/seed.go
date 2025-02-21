package seed

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/repository"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Seed struct {
	Wordlist *entities.Wordlist
}

var Wordlist = &Seed{
	Wordlist: &entities.Wordlist{
		UUID:                 uuid.New().String(),
		UserUUID:             uuid.New().String(), // we don't use it so we can fake them just for seeding
		ClientUUID:           uuid.New().String(), // we don't use it so we can fake them just for seeding
		WordlistName:         "rockyou.txt",
		WordlistHash:         "9076652d8ae75ce713e23ab09e10d9ee", // this is important and what actually what is checked
		WordlistSize:         139921497,
		WordlistLocationPath: constants.WordlistPath,
	},
}

func LoadWordlist(repo *repository.Repository) error {
	return loadWordlist(repo)
}

func loadWordlist(repo *repository.Repository) error {

	seeds := []*Seed{
		Wordlist,
	}

	for _, ww := range seeds {

		err := repo.InsertWordlist(ww.Wordlist)
		if err != nil {
			e := fmt.Errorf("failed to seed wordlist table: %v", err)
			log.Error(e)
			return e
		}
	}
	return nil
}
