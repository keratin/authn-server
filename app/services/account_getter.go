package services

import (
	"time"

	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/pkg/errors"
)

func AccountToJson(account *models.Account) map[string]interface{} {
	formattedLastLogin := ""
	if account.LastLoginAt != nil {
		formattedLastLogin = account.LastLoginAt.Format(time.RFC3339)
	}

	formattedPasswordChangedAt := ""
	if !account.PasswordChangedAt.IsZero() {
		formattedPasswordChangedAt = account.PasswordChangedAt.Format(time.RFC3339)
	}

	oAccountsMapped := []map[string]interface{}{}
	for _, oAccount := range account.OauthAccounts {
		oAccountsMapped = append(
			oAccountsMapped,
			map[string]interface{}{
				"provider":            oAccount.Provider,
				"provider_account_id": oAccount.ProviderID,
				"email":               oAccount.Email,
			},
		)
	}

	return map[string]interface{}{
		"id":                  account.ID,
		"username":            account.Username,
		"oauth_accounts":      oAccountsMapped,
		"last_login_at":       formattedLastLogin,
		"password_changed_at": formattedPasswordChangedAt,
		"locked":              account.Locked,
		"deleted":             account.DeletedAt != nil,
	}
}

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
