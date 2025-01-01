package environment

import (
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/utils"
)

type Environment struct {
	PCAPStorage        string
	HashcatFileStorage string
}

/*
InitEnvironment

- The function creates temporary directories where to save downloaded PCAP and converted ones
- Returns an Environment
*/
func InitEnvironment() (*Environment, error) {

	for _, dirs := range constants.ListOfDirToCreate {
		err := utils.CreateDirectory(dirs)
		if err != nil {
			return nil, err
		}
	}

	return &Environment{
		PCAPStorage:        constants.TempPCAPStorage,
		HashcatFileStorage: constants.TempHashcatFileDir,
	}, nil
}
