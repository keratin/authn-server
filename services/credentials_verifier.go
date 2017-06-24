package services

import (
	"database/sql"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"golang.org/x/crypto/bcrypt"
)

var emptyHashes = map[int]string{
	4:  "$2a$04$riUL94VEMOJwUfFkCUy8QO7HEL5L3uqUusOMELp509TuCWWJNuQG2",
	10: "$2a$10$1hP23Pl/f58gGNZeHHm80uqxrWUdALYVfp8aucGBmQiVRemEhZI7i",
	11: "$2a$11$GxV0LDD.xwM0ItzfbuMEDeMihmkIjs0Si6x6zhZtAAlm3p.6/3Z6q",
	12: "$2a$12$w58M3IGXURRAqXQ/OAsMmuqcV4YqP3WyJ.yHvHI5ANUK1bRWxeceK",
}

func CredentialsVerifier(store data.AccountStore, cfg *config.Config, username string, password string) (*models.Account, []Error) {
	account, err := store.FindByUsername(username)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	// if no account is found, we continue with a fake password hash. otherwise we
	// present a timing attack that can be used for user enumeration.
	var passwordHash []byte
	if err == sql.ErrNoRows {
		passwordHash = []byte(emptyHashes[cfg.BcryptCost])
	} else {
		passwordHash = []byte(account.Password)
	}

	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if err != nil {
		return nil, []Error{{"credentials", ErrFailed}}
	}
	if account.Locked {
		return nil, []Error{{"account", ErrLocked}}
	}
	if account.RequireNewPassword {
		return nil, []Error{{"credentials", ErrExpired}}
	}

	return account, nil
}
