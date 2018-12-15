package services

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
)

var bcryptPattern = regexp.MustCompile(`\A\$2[ayb]\$[0-9]{2}\$[A-Za-z0-9\.\/]{53}\z`)

func AccountImporter(store data.AccountStore, cfg *app.Config, username string, password string, locked bool) (*models.Account, error) {
	if username == "" {
		return nil, FieldErrors{{"username", ErrMissing}}
	}
	if password == "" {
		return nil, FieldErrors{{"password", ErrMissing}}
	}

	var hash []byte
	var err error
	if bcryptPattern.Match([]byte(password)) {
		hash = []byte(password)
	} else {
		hash, err = bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
		if err != nil {
			return nil, errors.Wrap(err, "bcrypt")
		}
	}

	acc, err := store.Create(username, hash)
	if err != nil {
		if data.IsUniquenessError(err) {
			return nil, FieldErrors{{"username", ErrTaken}}
		}

		return nil, errors.Wrap(err, "Create")
	}

	if locked {
		acc.Locked = true
		_, err := store.Lock(acc.ID)
		if err != nil {
			return nil, errors.Wrap(err, "Lock")
		}
	}

	return acc, nil
}
