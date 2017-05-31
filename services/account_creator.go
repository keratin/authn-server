package services

import (
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
	sqlite3 "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func AccountCreator(store data.AccountStore, cfg config.Config, username string, password string) (*data.Account, []Error) {
	errors := make([]Error, 0)

	if username == "" {
		errors = append(errors, Error{Field: "username", Message: ErrMissing})
	}
	if password == "" {
		errors = append(errors, Error{Field: "password", Message: ErrMissing})
	}

	if len(errors) > 0 {
		return nil, errors
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		panic(err)
	}

	acc, err := store.Create(username, hash)

	if err != nil {
		switch i := err.(type) {
		case sqlite3.Error:
			if i.ExtendedCode == sqlite3.ErrConstraintUnique {
				errors = append(errors, Error{Field: "username", Message: ErrTaken})
				return nil, errors
			} else {
				panic(err)
			}
		default:
			panic(err)
		}
	}

	return acc, nil
}
