package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/resets"
	"github.com/pkg/errors"
)

func PasswordResetSender(cfg *config.Config, account *models.Account) error {
	if cfg.AppPasswordResetURL == nil {
		return fmt.Errorf("AppPasswordResetURL unconfigured")
	}
	if account == nil {
		return FieldErrors{{"account", ErrMissing}}
	}
	if account.Locked {
		return FieldErrors{{"account", ErrLocked}}
	}

	reset, err := resets.New(cfg, account.ID, account.PasswordChangedAt)
	if err != nil {
		return errors.Wrap(err, "New Reset")
	}
	resetStr, err := reset.Sign(cfg.ResetSigningKey)
	if err != nil {
		return errors.Wrap(err, "Sign")
	}

	res, err := http.PostForm(cfg.AppPasswordResetURL.String(), url.Values{
		"account_id": []string{strconv.Itoa(account.ID)},
		"token":      []string{resetStr},
	})
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			// avoid reporting the URL with potential HTTP auth credentials
			return errors.Wrap(urlErr.Err, "PostForm")
		}
		return errors.Wrap(err, "PostForm")
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %v", res.StatusCode)
	}
	return nil
}
