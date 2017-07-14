package services

import (
	"fmt"
	"strings"

	"github.com/keratin/authn-server/config"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"
var ErrFormatInvalid = "FORMAT_INVALID"
var ErrInsecure = "INSECURE"
var ErrFailed = "FAILED"
var ErrLocked = "LOCKED"
var ErrExpired = "EXPIRED"
var ErrNotFound = "NOT_FOUND"
var ErrInvalidOrExpired = "INVALID_OR_EXPIRED"

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e fieldError) String() string {
	return fmt.Sprintf("%v: %v", e.Field, e.Message)
}

type FieldErrors []fieldError

func (es FieldErrors) Error() string {
	var buf = make([]string, len(es))
	for _, e := range es {
		buf = append(buf, e.String())
	}
	return strings.Join(buf, ", ")
}

func passwordValidator(cfg *config.Config, password string) *fieldError {
	if password == "" {
		return &fieldError{"password", ErrMissing}
	} else {
		// TODO: ensure that a super long password doesn't DOS the endpoint
		if zxcvbn.PasswordStrength(password, []string{}).Score < cfg.PasswordMinComplexity {
			return &fieldError{"password", ErrInsecure}
		}
	}
	return nil
}

func usernameValidator(cfg *config.Config, username string) *fieldError {
	if cfg.UsernameIsEmail {
		if isEmail(username) {
			if len(cfg.UsernameDomains) > 0 && !hasDomain(username, cfg.UsernameDomains) {
				return &fieldError{"username", ErrFormatInvalid}
			}
		} else {
			return &fieldError{"username", ErrFormatInvalid}
		}
	} else {
		if username == "" {
			return &fieldError{"username", ErrMissing}
		} else {
			if len(username) < cfg.UsernameMinLength {
				return &fieldError{"username", ErrFormatInvalid}
			}
		}
	}
	return nil
}
