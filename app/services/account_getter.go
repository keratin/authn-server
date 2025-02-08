package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
)

type AccountGetterParams struct {
	AccountID *int
	Username  *string
}

func AccountGetter(store data.AccountStore, params AccountGetterParams) (*models.Account, error) {
	var account *models.Account

	if params.AccountID != nil {
		ac, err := store.Find(*params.AccountID)
		if err != nil {
			return nil, errors.Wrap(err, "Find")
		}

		account = ac
	}

	if params.Username != nil && params.AccountID == nil {
		ac, err := store.FindByUsername(*params.Username)
		if err != nil {
			return nil, errors.Wrap(err, "FindByUsername")
		}

		account = ac
	}

	if account == nil {
		return nil, FieldErrors{{"account", ErrNotFound}}
	}

	oauthAccounts, err := store.GetOauthAccounts(account.ID)
	if err != nil {
		return nil, errors.Wrap(err, "GetOauthAccounts")
	}

	account.OauthAccounts = oauthAccounts
	return account, nil
}
