package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
)

func AccountGetter(store data.AccountStore, accountID int) (*models.Account, error) {
	account, err := store.Find(accountID)
	if err != nil {
		return nil, errors.Wrap(err, "Find")
	}
	if account == nil {
		return nil, FieldErrors{{"account", ErrNotFound}}
	}

	return account, nil
}
