package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
)

func AccountUpdater(store data.AccountStore, cfg *config.Config, accountID int, username string) error {
	username = strings.TrimSpace(username)

	fieldError := usernameValidator(cfg, username)
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
