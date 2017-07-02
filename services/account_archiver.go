package services

import "github.com/keratin/authn-server/data"

func AccountArchiver(store data.AccountStore, account_id int) []Error {
	account, err := store.Find(account_id)
	if err != nil {
		panic(err)
	}
	if account == nil {
		return []Error{Error{Field: "account", Message: ErrNotFound}}
	}

	store.Archive(account.Id)

	return nil
}
