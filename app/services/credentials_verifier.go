package services

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var emptyHashes = map[int]string{
	4:  "$2a$04$riUL94VEMOJwUfFkCUy8QO7HEL5L3uqUusOMELp509TuCWWJNuQG2",
	10: "$2a$10$1hP23Pl/f58gGNZeHHm80uqxrWUdALYVfp8aucGBmQiVRemEhZI7i",
	11: "$2a$11$GxV0LDD.xwM0ItzfbuMEDeMihmkIjs0Si6x6zhZtAAlm3p.6/3Z6q",
	12: "$2a$12$w58M3IGXURRAqXQ/OAsMmuqcV4YqP3WyJ.yHvHI5ANUK1bRWxeceK",
}

func CredentialsVerifier(store data.AccountStore, cfg *app.Config, username string, password string) (*models.Account, error) {
	if username == "" && password == "" {
		return nil, FieldErrors{{"credentials", ErrFailed}}
	}

	account, err := store.FindByUsername(username)
	if err != nil {
		return nil, errors.Wrap(err, "FindByUsername")
	}

	// if no account is found, we continue with a fake password hash. otherwise we
	// present a timing attack that can be used for user enumeration.
	var passwordHash []byte
	if account == nil {
		passwordHash = []byte(emptyHashes[cfg.BcryptCost])
	} else {
		passwordHash = []byte(account.Password)
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if account == nil || err != nil {
		return nil, FieldErrors{{"credentials", ErrFailed}}
	}
	if account.Locked {
		return nil, FieldErrors{{"account", ErrLocked}}
	}
	if account.RequireNewPassword {
		return nil, FieldErrors{{"credentials", ErrExpired}}
	}

	return account, nil
}
