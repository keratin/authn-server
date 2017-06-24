package models

import "time"

type Account struct {
	Id                 int
	Username           string
	Password           []byte
	Locked             bool
	RequireNewPassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          time.Time
}
