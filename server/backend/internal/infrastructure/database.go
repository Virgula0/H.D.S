package infrastructure

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
)

type Database struct {
	*sql.DB
}

// NewDatabaseConnection initializes a connection to the MariaDB database.
func NewDatabaseConnection() (*Database, error) {

	// Replace the connection parameters with your actual database connection details
	dbConnector, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		constants.DBUser,
		constants.DBPassword,
		constants.DBHost,
		constants.DBPort,
		constants.DBName,
	))

	if err != nil {
		return nil, err
	}

	// Check if the connection is actually valid by pinging the database
	if err := dbConnector.Ping(); err != nil {
		dbConnector.Close()
		return nil, err
	}

	return &Database{
		DB: dbConnector,
	}, nil
}

func (db *Database) DBPinger() {
	for {
		if err := db.DB.Ping(); err != nil {
			log.Fatalf("unable to connect to the database anymore %s", err.Error())
		}
		time.Sleep(time.Second * 10)
	}
}

func (db *Database) CloseDatabase() error {
	return db.DB.Close()
}

func (db *Database) CleanDB(tableNames []string) error {
	// wipe tables first, if requested
	cleanTables := make([]string, 0)

	for _, name := range tableNames {
		cleanTables = append(cleanTables, fmt.Sprintf("DELETE FROM %s", name))
	}

	for _, query := range cleanTables {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("unable to exec delete query %s ERROR: %v", query, err)
		}
	}
	return nil
}
