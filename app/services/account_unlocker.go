package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
)

func AccountUnlocker(store data.AccountStore, accountID int) error {
	affected, err := store.Unlock(accountID)
	if err != nil {
		return errors.Wrap(err, "Unlock")
	}
	if !affected {
		return FieldErrors{{"account", ErrNotFound}}
	}

	return nil
}
