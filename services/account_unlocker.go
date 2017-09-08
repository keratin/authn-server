package services

import "github.com/keratin/authn-server/data"

func AccountUnlocker(store data.AccountStore, accountID int) error {
	account, err := store.Find(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	store.Unlock(account.ID)

	return nil
}
