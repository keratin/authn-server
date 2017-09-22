package services

import (
	"net/url"
	"strconv"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func PasswordSetter(store data.AccountStore, r ops.ErrorReporter, cfg *config.Config, accountID int, password string) error {
	fieldError := passwordValidator(cfg, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return errors.Wrap(err, "GenerateFromPassword")
	}

	if cfg.AppPasswordChangedURL != nil {
		go func() {
			err := WebhookSender(cfg.AppPasswordChangedURL, &url.Values{
				"account_id": []string{strconv.Itoa(accountID)},
			})
			if err != nil {
				r.ReportError(err)
			}
		}()
	}

	return store.SetPassword(accountID, hash)
}
