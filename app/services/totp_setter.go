package services

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/compat"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

// TOTPSetter persists the OTP secret to the accountID if code is correct
func TOTPSetter(accountStore data.AccountStore, totpCache data.TOTPCache, cfg *app.Config, accountID int, code string) error {
	if code == "" { //Fail early if code is empty
		return FieldErrors{{"otp", ErrInvalidOrExpired}}
	}

	account, err := AccountGetter(accountStore, accountID)
	if err != nil {
		return err
	}

	secret, err := totpCache.LoadTOTPSecret(account.ID)
	if err != nil { //Error with redis itself
		return err
	}

	if !totp.Validate(code, string(secret)) { //Either cache expiry or validation error
		return FieldErrors{{"otp", ErrInvalidOrExpired}}
	}

	secret, err = compat.Encrypt(secret, cfg.DBEncryptionKey)
	if err != nil {
		return err
	}

	//Persist totp secret that was loaded from cache to db
	affected, err := accountStore.SetTOTPSecret(accountID, secret)
	if err != nil {
		return errors.Wrap(err, "TOTPSetter")
	}
	if !affected {
		return errors.New("unable to set totp secret")
	}

	// error here is not end of world it should timeout
	_ = totpCache.RemoveTOTPSecret(account.ID)

	return nil
}
