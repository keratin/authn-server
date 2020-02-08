package services

import (
	"fmt"
	"strings"

	"github.com/keratin/authn-server/app"
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

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e FieldError) String() string {
	return fmt.Sprintf("%v: %v", e.Field, e.Message)
}

func (e FieldError) Error() string {
	return e.String()
}

type FieldErrors []FieldError

func (es FieldErrors) Error() string {
	var buf = make([]string, len(es))
	for i, e := range es {
		buf[i] = e.Error()
	}
	return strings.Join(buf, ", ")
}

func PasswordValidator(cfg *app.Config, password string) *FieldError {
	if password == "" {
		return &FieldError{"password", ErrMissing}
	}

	score := CalcPasswordScore(password)

	if score < cfg.PasswordMinComplexity {
		return &FieldError{"password", ErrInsecure}
	}

	return nil
}

func UsernameValidator(cfg *app.Config, username string) *FieldError {
	if cfg.UsernameIsEmail {
		if !isEmail(username) {
			return &FieldError{"username", ErrFormatInvalid}
		}
		if len(cfg.UsernameDomains) > 0 && !hasDomain(username, cfg.UsernameDomains) {
			return &FieldError{"username", ErrFormatInvalid}
		}
	} else {
		if username == "" {
			return &FieldError{"username", ErrMissing}
		}
		if len(username) < cfg.UsernameMinLength {
			return &FieldError{"username", ErrFormatInvalid}
		}
	}
	return nil
}
