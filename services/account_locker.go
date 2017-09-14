package services

import (
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func AccountLocker(store data.AccountStore, accountID int) error {
	account, err := store.Find(accountID)
	if err != nil {
		return errors.Wrap(err, "Find")
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	return store.Lock(account.ID)
}
