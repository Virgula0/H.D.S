package environment

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

type Database struct {
	*sql.DB
}

// NewSQLiteConnection creates a connection to an SQLite database
func NewSQLiteConnection(dbPath string) (*Database, error) {
	dbConnector, err := sql.Open("sqlite3", fmt.Sprintf("%s&_foreign_keys=1&mode=rwc", dbPath))
	if err != nil {
		return nil, err
	}

	// Verify we can access the database file
	if err := dbConnector.Ping(); err != nil {
		dbConnector.Close()
		return nil, fmt.Errorf("failed to access SQLite database: %w", err)
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

// CreateTable dynamically constructs and executes a CREATE TABLE query
// it handles primary keys, foreign keys, and default values as annotations
// use the following syntax in structs
// `foreign:"user(UUID)"` | for foreign keys
// `primary:"true"` | for primary keys
// `default:"value"` | default value
// `db:"table_name"` | table name`
func (db *Database) CreateTable(tableName string, model any) error {
	t := reflect.TypeOf(model)

	var columns []string
	var primaryKey string
	var foreignKeys []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName := field.Tag.Get("db")
		dataType := getSQLiteType(field.Type)

		// Check constraints
		if field.Tag.Get("primary") == "true" {
			primaryKey = columnName
		}
		if fk := field.Tag.Get("foreign"); fk != "" {
			foreignKeys = append(foreignKeys, fmt.Sprintf("FOREIGN KEY(%s) REFERENCES %s ON DELETE CASCADE", columnName, fk))
		}
		if defaultValue := field.Tag.Get("default"); defaultValue != "" {
			dataType += " DEFAULT " + defaultValue
		}

		columns = append(columns, fmt.Sprintf("%s %s", columnName, dataType))
	}

	// Add primary key constraint
	if primaryKey != "" {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY(%s)", primaryKey))
	}

	// Combine all parts into the final query
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);", tableName, strings.Join(append(columns, foreignKeys...), ",\n  "))

	// Execute the query
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}

// Helper function to map Go types to SQLite types
func getSQLiteType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		return "TEXT"
	}
}
