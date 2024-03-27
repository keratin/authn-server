package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
)

func AccountOauthGetter(accountStore data.AccountStore, accountID int) ([]map[string]interface{}, error) {
	oAccountsMapped := []map[string]interface{}{}

	oauthAccounts, err := accountStore.GetOauthAccounts(accountID)
	if err != nil {
		return nil, errors.Wrap(err, "GetOauthAccounts")
	}

	for _, oAccount := range oauthAccounts {
		oAccountsMapped = append(
			oAccountsMapped,
			map[string]interface{}{
				"provider":    oAccount.Provider,
				"provider_id": oAccount.ProviderID,
			},
		)
	}

	return oAccountsMapped, nil
}
