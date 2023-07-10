package services

import (
	"net/url"
	"strconv"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func PasswordSetter(store data.AccountStore, r ops.ErrorReporter, cfg *app.Config, accountID int, password string) error {
	account, err := store.Find(accountID)
	if err != nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	fieldError := PasswordValidator(cfg, account.Username, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return errors.Wrap(err, "GenerateFromPassword")
	}

	affected, err := store.SetPassword(accountID, hash)
	if err != nil {
		return errors.Wrap(err, "SetPassword")
	}
	if !affected {
		return FieldErrors{{"account", ErrNotFound}}
	}

	if cfg.AppPasswordChangedURL != nil {
		go func() {
			err := WebhookSender(cfg.AppPasswordChangedURL, &url.Values{
				"account_id": []string{strconv.Itoa(accountID)},
			}, timeSensitiveDelivery, cfg.AppSigningKey)
			if err != nil {
				r.ReportError(err)
			}
		}()
	}

	return nil
}
