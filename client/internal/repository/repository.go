// #nosec G201 for SQL false positives
package repository

import (
	"database/sql"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
)

type Repository struct {
	*sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

// InsertWordlist creates a wordlist in db
func (repo *Repository) InsertWordlist(wordlist *entities.Wordlist) error {
	// User insert
	userQuery := fmt.Sprintf("INSERT INTO %s(uuid, uuid_user, client_uuid, wordlist_name, wordlist_hash, wordlist_lines) VALUES(?,?,?,?,?,?)", entities.WordlistTableName)
	if _, err := repo.Exec(userQuery, wordlist.UUID, wordlist.UserUUID, wordlist.ClientUUID, wordlist.WordlistName, wordlist.WordlistHash, wordlist.WordlistLines); err != nil {
		return fmt.Errorf("wordlist insert failed: %w", err)
	}

	return nil
}
