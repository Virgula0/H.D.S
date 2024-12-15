package seed

import (
	"crypto/md5"
	"fmt"
	"log"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/entities"
	"github.com/Virgula0/progetto-dp/server/backend/internal/repository"
	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
	"github.com/google/uuid"
)

type Seed struct {
	AdminUser *entities.User
	Role      constants.Role
}

var UserUUID = uuid.New().String()

var UserSeed = &Seed{
	// TODO: pay attention to this seed, in case the paassword is not random for debugging purposes
	AdminUser: &entities.User{UserUUID: UserUUID, Username: "admin", Password: "test1234"}, // utils.GenerateToken(32)},
	Role:      constants.ADMIN,
}

func LoadUsers(repo *repository.Repository) error {
	return loadUsers(repo)
}

func loadUsers(repo *repository.Repository) error {

	adminSeed := []*entities.User{
		UserSeed.AdminUser,
	}

	for _, user := range adminSeed {

		err := repo.CreateUser(user, constants.ADMIN)
		if err != nil {
			e := fmt.Errorf("failed to seed users table: %v", err)
			log.Println(e)
			return e
		}

		if constants.DebugEnabled != "" {
			randomHash := fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10))))

			_, err = repo.CreateClient(user.UserUUID, string(randomHash[:]), "127.0.0.1", "TestAdmin")

			if err != nil {
				e := fmt.Errorf("failed to seed client table: %v", err)
				log.Println(e)
				return e
			}

			_, err = repo.CerateRaspberryPI(user.UserUUID, fmt.Sprintf("%x", md5.Sum([]byte(utils.GenerateToken(10)))), fmt.Sprintf("%x%x", md5.Sum([]byte(utils.GenerateToken(10))), md5.Sum([]byte(utils.GenerateToken(10)))))

			if err != nil {
				e := fmt.Errorf("failed to seed rsp table: %v", err)
				log.Println(e)
				return e
			}
		}

	}
	return nil
}
