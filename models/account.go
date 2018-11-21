package models

import "time"

type Account struct {
	ID                 int
	Username           string
	Password           []byte
	Locked             bool
	RequireNewPassword bool       `db:"require_new_password"`
	PasswordChangedAt  time.Time  `db:"password_changed_at"`
	LastLoginAt        *time.Time `db:"last_login_at"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at"`
}

func (a Account) Archived() bool {
	return a.DeletedAt != nil
}
