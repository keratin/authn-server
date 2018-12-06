package services

import (
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func PasswordExpirer(store data.AccountStore, tokenStore data.RefreshTokenStore, accountID int) error {
	affected, err := store.RequireNewPassword(accountID)
	if err != nil {
		return errors.Wrap(err, "RequireNewPassword")
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
