package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

func AccountCreator(store data.AccountStore, cfg *config.Config, username string, password string) (*models.Account, []Error) {
	errors := make([]Error, 0)

	username = strings.TrimSpace(username)
	if cfg.UsernameIsEmail {
		if isEmail(username) {
			if len(cfg.UsernameDomains) > 0 && !hasDomain(username, cfg.UsernameDomains) {
				errors = append(errors, Error{Field: "username", Message: ErrFormatInvalid})
			}
		} else {
			errors = append(errors, Error{Field: "username", Message: ErrFormatInvalid})
		}
	} else {
		if username == "" {
			errors = append(errors, Error{Field: "username", Message: ErrMissing})
		} else {
			if len(username) < cfg.UsernameMinLength {
				errors = append(errors, Error{Field: "username", Message: ErrFormatInvalid})
			}
		}
	}

	if password == "" {
		errors = append(errors, Error{Field: "password", Message: ErrMissing})
	} else {
		if zxcvbn.PasswordStrength(password, []string{username}).Score < cfg.PasswordMinComplexity {
			errors = append(errors, Error{Field: "password", Message: ErrInsecure})
		}
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
		if data.IsUniquenessError(err) {
			errors = append(errors, Error{Field: "username", Message: ErrTaken})
			return nil, errors
		} else {
			panic(err)
		}
	}

	return acc, nil
}
