package services

import (
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"golang.org/x/crypto/bcrypt"
)

func PasswordSetter(store data.AccountStore, cfg *config.Config, accountId int, password string) error {
	fieldError := passwordValidator(cfg, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return err
	}

	err = store.SetPassword(accountId, hash)
	if err != nil {
		return err
	}
	return nil
}
