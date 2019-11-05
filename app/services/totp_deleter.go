package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
)

//TOTPDeleter removes TOTP from the specified account
func TOTPDeleter(accountStore data.AccountStore, accountID int) error {
	//Delete totp secret in database
	affected, err := accountStore.DeleteTOTPSecret(accountID)
	if err != nil {
		return errors.Wrap(err, "TOTPDeleter")
	}
	if !affected {
		return errors.New("unable to delete totp secret")
	}

	return nil
}
