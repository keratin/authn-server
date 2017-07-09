package services

import (
	"strings"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

func AccountCreator(store data.AccountStore, cfg *config.Config, username string, password string) (*models.Account, error) {
	errors := FieldErrors{}

	username = strings.TrimSpace(username)
	if cfg.UsernameIsEmail {
		if isEmail(username) {
			if len(cfg.UsernameDomains) > 0 && !hasDomain(username, cfg.UsernameDomains) {
				errors = append(errors, fieldError{"username", ErrFormatInvalid})
			}
		} else {
			errors = append(errors, fieldError{"username", ErrFormatInvalid})
		}
	} else {
		if username == "" {
			errors = append(errors, fieldError{"username", ErrMissing})
		} else {
			if len(username) < cfg.UsernameMinLength {
				errors = append(errors, fieldError{"username", ErrFormatInvalid})
			}
		}
	}

	if password == "" {
		errors = append(errors, fieldError{"password", ErrMissing})
	} else {
		if zxcvbn.PasswordStrength(password, []string{username}).Score < cfg.PasswordMinComplexity {
			errors = append(errors, fieldError{"password", ErrInsecure})
		}
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
