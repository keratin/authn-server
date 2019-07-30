package services

import (
	"net/url"
	"strconv"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func PasswordlessTokenSender(cfg *app.Config, account *models.Account, logger logrus.FieldLogger) error {
	if account == nil || account.Locked {
		return nil
	}

	passwordless, err := passwordless.New(cfg, account.ID)
	if err != nil {
		return errors.Wrap(err, "New Passwordless Token")
	}
	passwordlessStr, err := passwordless.Sign(cfg.PasswordlessTokenSigningKey)
	if err != nil {
		return errors.Wrap(err, "Sign")
	}

	err = WebhookSender(cfg.AppPasswordlessTokenURL, &url.Values{
		"account_id": []string{strconv.Itoa(account.ID)},
		"token":      []string{passwordlessStr},
	}, timeSensitiveDelivery)
	if err != nil {
		return errors.Wrap(err, "Webhook")
	}

	logger.WithField("accountID", account.ID).Info("sent passwordless token")

	return nil
}
