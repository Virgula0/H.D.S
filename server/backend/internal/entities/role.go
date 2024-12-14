package entities

const RoleTableName = "role"

type Role struct {
	RoleString string `db:"role_string"`
}
