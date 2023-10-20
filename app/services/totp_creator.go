package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/route"
	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var ErrExistingTOTPSecret = errors.New("a OTP secret has already been established for this account")

// TOTPCreator handles the creation and storage of new OTP tokens
func TOTPCreator(accountStore data.AccountStore, totpCache data.TOTPCache, accountID int, audience *route.Domain) (*otp.Key, error) {
	account, err := AccountGetter(accountStore, accountID)
	if err != nil {
		return nil, err
	}

	if account.TOTPEnabled() {
		// TODO: verify behavior here and test
		return nil, ErrExistingTOTPSecret
	}

	//Generate totp key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      audience.Hostname,
		AccountName: account.Username,
	})
	if err != nil {
		return nil, err
	}

	if err := totpCache.CacheTOTPSecret(account.ID, []byte(key.Secret())); err != nil {
		return nil, errors.Wrap(err, "TOTPCreator")
	}

	return key, nil
}
