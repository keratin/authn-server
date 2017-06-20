package services

import (
	"regexp"
	"strings"

	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
	"github.com/keratin/authn/models"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"
var ErrFormatInvalid = "FORMAT_INVALID"
var ErrInsecure = "INSECURE"

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

func hasDomain(email string, domains []string) bool {
	pieces := strings.SplitN(email, "@", 2)
	for _, domain := range domains {
		if domain == pieces[1] {
			return true
		}
	}
	return false
}

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
