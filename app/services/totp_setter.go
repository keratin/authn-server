package services

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/compat"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

//TOTPSetter persists the TOTP secret to the accountID if code is correct
func TOTPSetter(accountStore data.AccountStore, totpCache data.TOTPCache, cfg *app.Config, accountID int, code string) error {
	if code == "" { //Fail early if code is empty
		return FieldErrors{{"totp", ErrInvalidOrExpired}}
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
		return FieldErrors{{"totp", ErrInvalidOrExpired}}
	}

	secret, err = compat.Encrypt(secret, cfg.DBEncryptionKey)
	if err != nil {
		return err
	}

	//Persist totp secret that was loaded from cache to db TODO: Delete key from cache
	affected, err := accountStore.SetTOTPSecret(accountID, secret)
	if err != nil {
		return errors.Wrap(err, "TOTPSetter")
	}
	if !affected {
		return errors.New("unable to set totp secret")
	}

	return nil
}
