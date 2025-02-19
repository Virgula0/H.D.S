package environment

import (
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/entities"
	"github.com/Virgula0/progetto-dp/client/internal/repository"
	"github.com/Virgula0/progetto-dp/client/internal/usecase"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	"path/filepath"
)

type Environment struct {
	PCAPStorage        string
	HashcatFileStorage string
	Keys               *utils.Keys
}

type ServiceHandler struct {
	Usecase *usecase.Usecase
}

type Table struct {
	Entity    any
	TableName string
}

var tables = []*Table{
	{
		Entity:    entities.Wordlist{},
		TableName: "wordlist",
	},
}

func (db *Database) initDB() error {

	var tableNames []string
	for _, table := range tables {
		if err := db.CreateTable(table.TableName, table.Entity); err != nil {
			return err
		}

		tableNames = append(tableNames, table.TableName)
	}

	if constants.WipeTables {
		// Delete data from DB
		if err := db.CleanDB(tableNames); err != nil {
			return err
		}
	}

	go db.StartDBPinger()

	return nil
}

/*
InitEnvironment

- The function creates temporary directories where to save downloaded PCAP and converted ones
- Returns an Environment
*/
func InitEnvironment() (*Environment, ServiceHandler, error) {

	// create temp dirs
	for _, dirs := range constants.ListOfDirToCreate {
		err := utils.CreateDirectory(dirs)
		if err != nil {
			return nil, ServiceHandler{}, err
		}
	}

	var keys = &utils.Keys{}

	rec := utils.CallBackFunc{
		CallBack: ClassifyFile,
		Keys:     keys,
	}

	// read files within certs recursively
	if err := filepath.WalkDir(constants.CertFileDir, rec.RecursiveDirectoryWalk); err != nil {
		return nil, ServiceHandler{}, err
	}

	// init db
	db, err := NewSQLiteConnection("file::memory:?cache=shared") // in memory database

	if err != nil {
		return nil, ServiceHandler{}, err
	}

	repo := repository.NewRepository(db.DB) // init repository

	if err := db.initDB(); err != nil {
		return nil, ServiceHandler{}, err
	}

	return &Environment{
			PCAPStorage:        constants.TempPCAPStorage,
			HashcatFileStorage: constants.TempHashcatFileDir,
			Keys:               keys,
		},
		ServiceHandler{Usecase: usecase.NewUsecase(repo)},
		nil
}

// EmptyCerts check if certs have been imported
func (e *Environment) EmptyCerts() bool {
	if e.Keys.ClientCert == nil ||
		e.Keys.ClientKey == nil ||
		e.Keys.CACert == nil {
		return true
	}
	return false
}
