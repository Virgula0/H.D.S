package usecase

import (
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/repository"
)

type Usecase struct {
	repo *repository.Repository
}

func NewUsecase(repo *repository.Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (uc *Usecase) CreateWordlist(wordlist *entities.Wordlist) error {
	return uc.repo.InsertWordlist(wordlist)
}

func (uc *Usecase) GetWordlistByHash(hash string) (*entities.Wordlist, error) {
	return uc.repo.GetWordlistByHash(hash)
}
