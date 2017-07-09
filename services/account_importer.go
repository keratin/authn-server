package services

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
)

var bcryptPattern = regexp.MustCompile(`\A\$2[ayb]\$[0-9]{2}\$[A-Za-z0-9\.\/]{53}\z`)

func AccountImporter(store data.AccountStore, cfg *config.Config, username string, password string, locked bool) (*models.Account, []Error) {
	if username == "" {
		return nil, []Error{{"username", ErrMissing}}
	}
	if password == "" {
		return nil, []Error{{"password", ErrMissing}}
	}

	var hash []byte
	var err error
	if bcryptPattern.Match([]byte(password)) {
		hash = []byte(password)
	} else {
		hash, err = bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
		if err != nil {
			panic(err)
		}
	}

	acc, err := store.Create(username, hash)
	if err != nil {
		if data.IsUniquenessError(err) {
			return nil, []Error{{"username", ErrTaken}}
		} else {
			panic(err)
		}
	}

	if locked {
		acc.Locked = true
		err = store.Lock(acc.Id)
		if err != nil {
			panic(err)
		}
	}

	return acc, nil
}
