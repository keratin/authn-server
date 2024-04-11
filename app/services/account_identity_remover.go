package services

import (
	"github.com/keratin/authn-server/app/data"
)

func IdentityRemover(store data.AccountStore, accountId int, providers []string) error {
	account, err := store.Find(accountId)
	if err != nil {
		return err
	}

	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	for _, provider := range providers {
		_, err = store.DeleteOauthAccount(accountId, provider)
		if err != nil {
			return err
		}
	}

	return nil
}
