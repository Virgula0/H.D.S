package repository

import "database/sql"

type Repository struct {
	*sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}
