package data

import "github.com/keratin/authn-server/models"

type AccountStore interface {
	Create(u string, p []byte) (*models.Account, error)
	Find(id int) (*models.Account, error)
	FindByUsername(u string) (*models.Account, error)
	Archive(id int) error
	Lock(id int) error
	Unlock(id int) error
	RequireNewPassword(id int) error
	SetPassword(id int, p []byte) error
}
