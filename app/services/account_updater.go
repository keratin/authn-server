package services

import (
	"strings"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
)

func AccountUpdater(store data.AccountStore, cfg *app.Config, accountID int, username string) error {
	username = strings.TrimSpace(username)

	fieldError := UsernameValidator(cfg, username)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	affected, err := store.UpdateUsername(accountID, username)
	if err != nil {
		if data.IsUniquenessError(err) {
			return FieldErrors{{"username", ErrTaken}}
		}

		return errors.Wrap(err, "UpdateUsername")
	}
	if !affected {
		return FieldErrors{{"account", ErrNotFound}}
	}

	return nil
}
