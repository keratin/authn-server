package services

import (
	"github.com/keratin/authn-server/data"
)

func LastLoginUpdater(store data.AccountStore, accountID int) error {
	rowsIsAffected, err := store.SetLastLogin(accountID)
	if rowsIsAffected == false {
		return FieldErrors{{"account", ErrNotFound}}
	}

	return err
}
