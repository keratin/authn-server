package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountUnlocker(t *testing.T) {
	store := mock.NewAccountStore()

	lockedAccount, err := store.Create("locked@keratin.tech", []byte("password"))
	require.NoError(t, err)
	_, err = store.Lock(lockedAccount.ID)
	require.NoError(t, err)

	unlockedAccount, err := store.Create("unlocked@keratin.tech", []byte("password"))
	require.NoError(t, err)

	var testCases = []struct {
		accountID int
		errors    *services.FieldErrors
	}{
		{123456789, &services.FieldErrors{{"account", services.ErrNotFound}}},
		{lockedAccount.ID, nil},
		{unlockedAccount.ID, nil},
	}

	for _, tc := range testCases {
		errs := services.AccountUnlocker(store, tc.accountID)
		if tc.errors == nil {
			assert.Empty(t, errs)
			acct, err := store.Find(tc.accountID)
			require.NoError(t, err)
			assert.False(t, acct.Locked)
		} else {
			assert.Equal(t, *tc.errors, errs)
		}
	}
}
