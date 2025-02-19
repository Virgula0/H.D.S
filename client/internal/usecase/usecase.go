package usecase

import (
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
