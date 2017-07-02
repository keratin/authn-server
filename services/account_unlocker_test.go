package services_test

import (
	"testing"

	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountUnlocker(t *testing.T) {
	store := mock.NewAccountStore()

	locked_account, err := store.Create("locked@keratin.tech", []byte("password"))
	require.NoError(t, err)
	err = store.Lock(locked_account.Id)
	require.NoError(t, err)

	unlocked_account, err := store.Create("unlocked@keratin.tech", []byte("password"))
	require.NoError(t, err)

	var testCases = []struct {
		account_id int
		errors     *[]services.Error
	}{
		{123456789, &[]services.Error{{"account", services.ErrNotFound}}},
		{locked_account.Id, nil},
		{unlocked_account.Id, nil},
	}

	for _, tc := range testCases {
		errs := services.AccountUnlocker(store, tc.account_id)
		if tc.errors == nil {
			assert.Empty(t, errs)
			acct, err := store.Find(tc.account_id)
			require.NoError(t, err)
			assert.False(t, acct.Locked)
		} else {
			assert.Equal(t, *tc.errors, errs)
		}
	}
}
