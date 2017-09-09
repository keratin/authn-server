package services

import (
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func PasswordSetter(store data.AccountStore, cfg *config.Config, accountID int, password string) error {
	fieldError := passwordValidator(cfg, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return errors.Wrap(err, "GenerateFromPassword")
	}

	err = store.SetPassword(accountID, hash)
	if err != nil {
		return errors.Wrap(err, "SetPassword")
	}
	return nil
}
