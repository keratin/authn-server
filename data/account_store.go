package data

import "github.com/keratin/authn-server/models"

type AccountStore interface {
	Create(u string, p []byte) (*models.Account, error)
	FindByUsername(u string) (*models.Account, error)
}
