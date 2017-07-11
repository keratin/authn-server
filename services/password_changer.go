package services

import (
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"golang.org/x/crypto/bcrypt"
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

	fieldError := passwordValidator(cfg, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return err
	}

	err = store.SetPassword(id, hash)
	if err != nil {
		return err
	}
	return nil
}
