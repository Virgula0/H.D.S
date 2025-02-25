// #nosec G201 for SQL false positives
package repository

import (
	"database/sql"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/customerrors"
	"github.com/Virgula0/progetto-dp/client/internal/entities"

	log "github.com/sirupsen/logrus"
)

type Repository struct {
	*sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func scanRowsInterfaceWrapper(rows *sql.Rows, builder func() (any, []any)) ([]any, error) {
	defer rows.Close()
	var results []any

	for rows.Next() {
		entity, dest := builder()
		if err := rows.Scan(dest...); err != nil {
			log.Error("Row scan error: ", err.Error())
			return nil, customerrors.ErrInternalServerError
		}
		results = append(results, entity)
	}

	if err := rows.Err(); err != nil {
		log.Error("Rows iteration error: ", err.Error())
		return nil, customerrors.ErrInternalServerError
	}
	return results, nil
}

func (repo *Repository) queryEntities(query string, builder func() (any, []any), args ...any) ([]any, error) {
	rows, err := repo.Query(query, args...)
	if err != nil {
		log.Error("Query execution error: ", err.Error())
		return nil, customerrors.ErrInternalServerError
	}

	// Use type-asserted scanRows wrapper to handle interface conversion
	results, err := scanRowsInterfaceWrapper(rows, builder)
	if err != nil {
		return nil, err
	}

	return results, nil
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
