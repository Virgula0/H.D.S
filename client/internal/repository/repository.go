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

// InsertWordlist creates a wordlist in the database
func (repo *Repository) InsertWordlist(wordlist *entities.Wordlist) error {
	// Wordlist insert query
	userQuery := fmt.Sprintf(
		"INSERT INTO %s(uuid, uuid_user, client_uuid, wordlist_name, wordlist_hash, wordlist_size, wordlist_location_path) VALUES(?,?,?,?,?,?,?)",
		entities.WordlistTableName,
	)

	_, err := repo.Exec(userQuery, wordlist.UUID, wordlist.UserUUID, wordlist.ClientUUID, wordlist.WordlistName, wordlist.WordlistHash, wordlist.WordlistSize, wordlist.WordlistLocationPath)
	if err != nil {
		return err
	}

	return nil
}

// GetWordlistByHash creates a wordlist in the database
func (repo *Repository) GetWordlistByHash(hash string) (*entities.Wordlist, error) {
	var wordlist entities.Wordlist

	query := fmt.Sprintf("SELECT * FROM %s WHERE wordlist_hash = ?", entities.WordlistTableName)

	row := repo.QueryRow(query, hash)
	err := row.Scan(&wordlist.UUID, &wordlist.UserUUID, &wordlist.ClientUUID, &wordlist.WordlistName, &wordlist.WordlistHash, &wordlist.WordlistSize, &wordlist.WordlistLocationPath)

	return &wordlist, err
}
