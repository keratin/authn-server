package services

import (
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func AccountLocker(store data.AccountStore, tokenStore data.RefreshTokenStore, accountID int) error {
	affected, err := store.Lock(accountID)
	if err != nil {
		return errors.Wrap(err, "Lock")
	}
	if !affected {
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

	return nil
}
