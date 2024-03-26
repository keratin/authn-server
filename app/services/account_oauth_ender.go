package services

import (
	"math"

	"github.com/keratin/authn-server/app/data"
)

func AccountOauthEnder(store data.AccountStore, accountId int, provider string) error {
	account, err := store.Find(accountId)
	if err != nil {
		return err
	}

	oauthAccounts, err := store.GetOauthAccounts(accountId)
	if err != nil {
		return err
	}

	for _, oAccount := range oauthAccounts {
		if math.Abs(oAccount.CreatedAt.Sub(account.PasswordChangedAt).Seconds()) < 5 {
			return FieldErrors{{Field: "password", Message: ErrPasswordResetRequired}}
		}
	}

	_, err = store.DeleteOauthAccount(accountId, provider)
	if err != nil {
		return err
	}

	return nil
}
