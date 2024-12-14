package entities

const UserTableName = "user"

type User struct {
	UserUUID string `db:"uuid"`
	Username string `db:"username"`
	Password string `db:"password"`
}
