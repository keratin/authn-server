package services

import "github.com/keratin/authn-server/data"

func PasswordExpirer(store data.AccountStore, tokenStore data.RefreshTokenStore, accountID int) error {
	account, err := store.Find(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	}

	err = store.RequireNewPassword(accountID)
	if err != nil {
		return err
	}

	tokens, err := tokenStore.FindAll(accountID)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		err = tokenStore.Revoke(token)
		if err != nil {
			return err
		}
	}

	return nil
}
