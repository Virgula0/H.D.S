package seed

import (
	"crypto/md5" // #nosec G501 disable weak hash alert, it is not used for crypto stuff
	"fmt"
	"log"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/google/uuid"
)

type Seed struct {
	User *entities.User
	Role constants.Role
}

var AdminSeed = &Seed{
	// TODO: pay attention to this seed, in case the paassword is not random for debugging purposes
	User: &entities.User{UserUUID: uuid.New().String(), Username: "admin", Password: "test1234"}, // utils.GenerateToken(32)},
	Role: constants.ADMIN,
}

var NormalUserSeed = &Seed{
	// TODO: pay attention to this seed, in case the paassword is not random for debugging purposes
	User: &entities.User{UserUUID: uuid.New().String(), Username: "user", Password: "test1234"}, // utils.GenerateToken(32)},
	Role: constants.USER,
}

func LoadUsers(repo *repository.Repository) error {
	return loadUsers(repo)
}

func loadUsers(repo *repository.Repository) error {

	adminSeed := []*entities.User{
		AdminSeed.User,
		NormalUserSeed.User,
	}

	for _, user := range adminSeed {

		err := repo.CreateUser(user, constants.ADMIN)
		if err != nil {
			e := fmt.Errorf("failed to seed users table: %v", err)
			log.Println(e)
			return e
		}

		if constants.DebugEnabled != "" {
			randomHash := fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))) // #nosec G401 disable weak hash alert, it is not used for crypto stuff

			_, err = repo.CreateClient(user.UserUUID, randomHash, "127.0.0.1", "TestAdmin")

			if err != nil {
				e := fmt.Errorf("failed to seed client table: %v", err)
				log.Println(e)
				return e
			}

			// #nosec G401 disable weak hash alert, it is not used for crypto stuff
			rspID, err := repo.CreateRaspberryPI(user.UserUUID, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))), fmt.Sprintf("%x%x", md5.Sum([]byte(utils.GenerateToken(10))), md5.Sum([]byte(utils.GenerateToken(10)))))

			if err != nil {
				e := fmt.Errorf("failed to seed rsp table: %v", err)
				log.Println(e)
				return e
			}

			_, err = repo.CreateHandshake(user.UserUUID, rspID, "TEST", "XX:XX:XX:XX:XX:XX", "Uncracked", utils.StringToBase64String("test.pcap"))

			if err != nil {
				e := fmt.Errorf("failed to seed handshake table: %v", err)
				log.Println(e)
				return e
			}
		}

	}
	return nil
}
