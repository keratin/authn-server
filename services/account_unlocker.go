package services

import "github.com/keratin/authn-server/data"

func AccountUnlocker(store data.AccountStore, accountId int) error {
	account, err := store.Find(accountId)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	store.Unlock(account.Id)

	return nil
}
