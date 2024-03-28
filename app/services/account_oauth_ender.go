package services

import (
	"math"
	"strings"

	"github.com/keratin/authn-server/app/data"
)

func AccountOauthEnder(store data.AccountStore, accountId int, providers []string) error {
	account, err := store.Find(accountId)
	if err != nil {
		return err
	}

	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	oauthAccounts, err := store.GetOauthAccounts(accountId)
	if err != nil {
		return err
	}

	mappedProviders := map[string]uint8{}
	for _, provider := range providers {
		mappedProviders[strings.ToLower(provider)] = 1
	}

	for _, oAccount := range oauthAccounts {
		_, isProviderMatched := mappedProviders[strings.ToLower(oAccount.Provider)]
		hasRandomOauthPassword := math.Abs(oAccount.CreatedAt.Sub(account.PasswordChangedAt).Seconds()) < 5

		if hasRandomOauthPassword && isProviderMatched {
			return FieldErrors{{"password", ErrNewPasswordRequired}}
		}
	}

	for _, provider := range providers {
		_, err = store.DeleteOauthAccount(accountId, provider)
		if err != nil {
			return err
		}
	}

	return nil
}
