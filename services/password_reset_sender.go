package services

import (
	"net/url"
	"strconv"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/resets"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func PasswordResetSender(cfg *config.Config, account *models.Account) error {
	if account == nil || account.Locked {
		return nil
	}

	reset, err := resets.New(cfg, account.ID, account.PasswordChangedAt)
	if err != nil {
		return errors.Wrap(err, "New Reset")
	}
	resetStr, err := reset.Sign(cfg.ResetSigningKey)
	if err != nil {
		return errors.Wrap(err, "Sign")
	}

	err = WebhookSender(cfg.AppPasswordResetURL, &url.Values{
		"account_id": []string{strconv.Itoa(account.ID)},
		"token":      []string{resetStr},
	})
	if err != nil {
		return errors.Wrap(err, "Webhook")
	}

	log.WithFields(log.Fields{"accountID": account.ID}).Info("sent password reset token")

	return nil
}
