package services

import (
	"fmt"
	"strings"

	"github.com/keratin/authn-server/app"
)

var (
	ErrMissing               = "MISSING"
	ErrTaken                 = "TAKEN"
	ErrFormatInvalid         = "FORMAT_INVALID"
	ErrInsecure              = "INSECURE"
	ErrFailed                = "FAILED"
	ErrLocked                = "LOCKED"
	ErrExpired               = "EXPIRED"
	ErrNotFound              = "NOT_FOUND"
	ErrInvalidOrExpired      = "INVALID_OR_EXPIRED"
	ErrPasswordResetRequired = "PASSWORD_RESET_REQUIRED"
)

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
	buf := make([]string, len(es))
	for i, e := range es {
		buf[i] = e.Error()
	}
	return strings.Join(buf, ", ")
}

func PasswordValidator(cfg *app.Config, username, password string) *FieldError {
	if password == "" {
		return &FieldError{"password", ErrMissing}
	}

	if username == password {
		return &FieldError{"password", ErrInsecure}
	}

	score := CalculatePasswordScore(password)

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
