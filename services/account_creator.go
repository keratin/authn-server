package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"golang.org/x/crypto/bcrypt"
)

func AccountCreator(store data.AccountStore, cfg *config.Config, username string, password string) (*models.Account, error) {
	username = strings.TrimSpace(username)

	errors := FieldErrors{}

	fieldError := usernameValidator(cfg, username)
	if fieldError != nil {
		errors = append(errors, *fieldError)
	}

	fieldError = passwordValidator(cfg, password)
	if fieldError != nil {
		errors = append(errors, *fieldError)
	}

	if len(errors) > 0 {
		return nil, errors
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	acc, err := store.Create(username, hash)

	if err != nil {
		if data.IsUniquenessError(err) {
			return nil, FieldErrors{{"username", ErrTaken}}
		} else {
			return nil, err
		}
	}

	return acc, nil
}
