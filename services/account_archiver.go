package services

import "github.com/keratin/authn-server/data"

func AccountArchiver(store data.AccountStore, accountID int) error {
	account, err := store.Find(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	store.Archive(account.ID)

	return nil
}
