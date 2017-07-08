package services

import "github.com/keratin/authn-server/data"

func PasswordExpirer(store data.AccountStore, tokenStore data.RefreshTokenStore, account_id int) []Error {
	account, err := store.Find(account_id)
	if err != nil {
		panic(err)
	}
	if account == nil {
		return []Error{Error{Field: "account", Message: ErrNotFound}}
	}

	err = store.RequireNewPassword(account_id)
	if err != nil {
		panic(err)
	}

	tokens, err := tokenStore.FindAll(account_id)
	if err != nil {
		panic(err)
	}
	for _, token := range tokens {
		err = tokenStore.Revoke(token)
		if err != nil {
			panic(err)
		}
	}

	return nil
}
