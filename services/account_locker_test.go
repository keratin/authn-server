package services_test

import (
	"testing"

	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLocker(t *testing.T) {
	store := mock.NewAccountStore()

	lockedAccount, err := store.Create("locked@keratin.tech", []byte("password"))
	require.NoError(t, err)
	err = store.Lock(lockedAccount.Id)
	require.NoError(t, err)

	unlockedAccount, err := store.Create("unlocked@keratin.tech", []byte("password"))
	require.NoError(t, err)

	var testCases = []struct {
		accountId int
		errors    *services.FieldErrors
	}{
		{123456789, &services.FieldErrors{{"account", services.ErrNotFound}}},
		{lockedAccount.Id, nil},
		{unlockedAccount.Id, nil},
	}

	for _, tc := range testCases {
		errs := services.AccountLocker(store, tc.accountId)
		if tc.errors == nil {
			assert.Empty(t, errs)
			acct, err := store.Find(tc.accountId)
			require.NoError(t, err)
			assert.True(t, acct.Locked)
		} else {
			assert.Equal(t, *tc.errors, errs)
		}
	}
}
