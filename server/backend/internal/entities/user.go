package entities

const UserTableName = "user"

type User struct {
	UserUUID string `db:"UUID"`
	Username string `db:"USERNAME"`
	Password string `db:"PASSWORD"`
}
