package services

import "github.com/keratin/authn-server/data"

func AccountArchiver(store data.AccountStore, account_id int) error {
	account, err := store.Find(account_id)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	store.Archive(account.Id)

	return nil
}
