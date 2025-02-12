package entities

const UserTableName = "user"

type User struct {
	UserUUID string `db:"UUID"`
	Username string `db:"USERNAME"`
	Password string `db:"PASSWORD"`
}

type UpdateUserPasswordRequest struct {
	OldPassword        string `json:"oldPassword" validate:"required"`
	NewPassword        string `json:"newPassword" validate:"required"`
	NewPasswordConfirm string `json:"newPasswordConfirm" validate:"required"`
}

type UpdateUserPasswordResponse struct {
	Status string `json:"status"`
}
