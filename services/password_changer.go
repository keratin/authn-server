package services

import (
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
)

func PasswordChanger(store data.AccountStore, cfg *config.Config, id int, password string) error {
	account, err := store.Find(id)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	} else if account.Locked {
		return FieldErrors{{"account", ErrLocked}}
	}

	return PasswordSetter(store, cfg, id, password)
}
