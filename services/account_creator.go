package services

import (
	"regexp"
	"strings"

	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
	sqlite3 "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"
var ErrFormatInvalid = "FORMAT_INVALID"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// worried about an imperfect regex? see: http://www.regular-expressions.info/email.html
var emailPattern = regexp.MustCompile(`(?i)\A[A-Z0-9._%+-]*@(?:[A-Z0-9-]*\.)*[A-Z]*\z`)

func isEmail(s string) bool {
	// SECURITY: the len() check prevents a regex ddos via overly large usernames
	return len(s) < 255 && emailPattern.Match([]byte(s))
}

func hasDomain(email string, domain string) bool {
	pieces := strings.SplitN(email, "@", 2)
	return domain == pieces[1]
}

func AccountCreator(store data.AccountStore, cfg *config.Config, username string, password string) (*data.Account, []Error) {
	errors := make([]Error, 0)

	username = strings.TrimSpace(username)
	if cfg.UsernameIsEmail {
		if isEmail(username) {
			if len(cfg.UsernameDomain) > 0 && !hasDomain(username, cfg.UsernameDomain) {
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
