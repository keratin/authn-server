package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
)

func AccountArchiver(store data.AccountStore, tokenStore data.RefreshTokenStore, accountID int) error {
	affected, err := store.Archive(accountID)
	if err != nil {
		return errors.Wrap(err, "Archive")
	}
	if !affected {
		return FieldErrors{{"account", ErrNotFound}}
	}

	return SessionBatchEnder(tokenStore, accountID)
}
