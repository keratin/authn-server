package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
)

func AccountUpdater(store data.AccountStore, cfg *config.Config, id int, username string) error {
	username = strings.TrimSpace(username)

	fieldError := usernameValidator(cfg, username)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	return store.UpdateUsername(id, username)
}
