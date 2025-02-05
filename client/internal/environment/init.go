package environment

import (
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
	"path/filepath"
)

type Environment struct {
	PCAPStorage        string
	HashcatFileStorage string
	Keys               *utils.Keys
}

/*
InitEnvironment

- The function creates temporary directories where to save downloaded PCAP and converted ones
- Returns an Environment
*/
func InitEnvironment() (*Environment, error) {

	// create temp dirs
	for _, dirs := range constants.ListOfDirToCreate {
		err := utils.CreateDirectory(dirs)
		if err != nil {
			return nil, err
		}
	}

	var keys = &utils.Keys{}

	rec := utils.CallBackFunc{
		CallBack: ClassifyFile,
		Keys:     keys,
	}

	// read files within certs recursively
	if err := filepath.WalkDir(constants.CertFileDir, rec.RecursiveDirectoryWalk); err != nil {
		return nil, err
	}

	return &Environment{
		PCAPStorage:        constants.TempPCAPStorage,
		HashcatFileStorage: constants.TempHashcatFileDir,
		Keys:               keys,
	}, nil
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
