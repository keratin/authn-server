package services

import (
	"github.com/keratin/authn-server/app/data"
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

	return SessionBatchEnder(tokenStore, accountID)
}
