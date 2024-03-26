package services

import (
	"math"

	"github.com/keratin/authn-server/app/data"
)

type AccountOauthEnderResult struct {
	RequirePasswordReset bool
}

func AccountOauthEnder(store data.AccountStore, accountId int, provider string) (*AccountOauthEnderResult, error) {
	result := &AccountOauthEnderResult{}

	account, err := store.Find(accountId)
	if err != nil {
		return nil, err
	}

	oauthAccounts, err := store.GetOauthAccounts(accountId)
	if err != nil {
		return nil, err
	}

	for _, oAccount := range oauthAccounts {
		if math.Abs(oAccount.CreatedAt.Sub(account.PasswordChangedAt).Seconds()) < 5 {
			result.RequirePasswordReset = true
			store.RequireNewPassword(accountId)
		}
	}

	_, err = store.DeleteOauthAccount(accountId, provider)
	if err != nil {
		return nil, err
	}

	return result, nil
}
