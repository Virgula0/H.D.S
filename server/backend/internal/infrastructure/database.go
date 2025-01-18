package infrastructure

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type Database struct {
	*sql.DB
}

// NewDatabaseConnection initializes a connection to the MariaDB database.
func NewDatabaseConnection(dbUser, dbPassword, dbHost, dbPort, dbName string) (*Database, error) {

	dbConnector, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
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

func (db *Database) dbPinger() error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		if err := db.DB.Ping(); err != nil {
			return err
		}
		<-ticker.C
	}
}

func (db *Database) StartDBPinger() {
	errDB := db.dbPinger()
	if errDB != nil {
		log.Fatalf("Unable to connect to the database anymore %s", errDB.Error())
	}
}

func (db *Database) CloseDatabase() error {
	return db.DB.Close()
}

// CleanDB utility function (called from teardown tests, for cleaning tables)
func (db *Database) CleanDB(tableNames []string) error {
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
