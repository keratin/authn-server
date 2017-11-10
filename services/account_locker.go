package services

import (
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func AccountLocker(store data.AccountStore, tokenStore data.RefreshTokenStore, accountID int) error {
	account, err := store.Find(accountID)
	if err != nil {
		return errors.Wrap(err, "Find")
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	tokens, err := tokenStore.FindAll(accountID)
	if err != nil {
		return errors.Wrap(err, "FindAll")
	}
	for _, token := range tokens {
		err = tokenStore.Revoke(token)
		if err != nil {
			return errors.Wrap(err, "Revoke")
		}
	}

	return store.Lock(account.ID)
}
