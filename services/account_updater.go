package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func AccountUpdater(store data.AccountStore, cfg *config.Config, accountID int, username string) error {
	account, err := store.Find(accountID)
	if err != nil {
		return errors.Wrap(err, "Find")
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	username = strings.TrimSpace(username)

	fieldError := usernameValidator(cfg, username)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	err = store.UpdateUsername(accountID, username)
	if err != nil {
		if data.IsUniquenessError(err) {
			return FieldErrors{{"username", ErrTaken}}
		}

		return errors.Wrap(err, "UpdateUsername")
	}
	return nil
}
