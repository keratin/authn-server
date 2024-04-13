package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
)

func AccountGetter(store data.AccountStore, accountID int) (*models.Account, error) {
	account, err := store.Find(accountID)
	if err != nil {
		return nil, errors.Wrap(err, "Find")
	}
	if account == nil {
		return nil, FieldErrors{{"account", ErrNotFound}}
	}

	oauthAccounts, err := store.GetOauthAccounts(accountID)
	if err != nil {
		return nil, errors.Wrap(err, "GetOauthAccounts")
	}

	account.OauthAccounts = oauthAccounts
	return account, nil
}
