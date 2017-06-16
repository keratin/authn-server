package data

import "github.com/keratin/authn/models"

type AccountStore interface {
	Create(u string, p []byte) (*models.Account, error)
}
