package services

import (
	"strconv"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/tokens/password_resets"
)

func PasswordResetter(store data.AccountStore, cfg *config.Config, token string, password string) (int, error) {
	claims, err := password_resets.Parse(token, cfg)
	if err != nil {
		return 0, FieldErrors{{"token", ErrInvalidOrExpired}}
	}

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, err
	}

	account, err := store.Find(id)
	if err != nil {
		return 0, err
	}
	if account == nil {
		return 0, FieldErrors{{"account", ErrNotFound}}
	} else if account.Locked {
		return 0, FieldErrors{{"account", ErrLocked}}
	} else if account.Archived() {
		return 0, FieldErrors{{"account", ErrLocked}}
	}

	if claims.LockExpired(account.PasswordChangedAt) {
		return 0, FieldErrors{{"token", ErrInvalidOrExpired}}
	}

	return account.ID, PasswordSetter(store, cfg, id, password)
}
