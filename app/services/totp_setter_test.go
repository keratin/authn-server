package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPSetter(t *testing.T) {
	accountStore := mock.NewAccountStore(mock.WithSetTOTPFailureID(1))
	// create first so it gets the "failure ID" of 1
	failSetAccount, err := accountStore.Create("bad secret user", []byte("password"))
	require.NoError(t, err)
	account, err := accountStore.Create("test user", []byte("password"))
	require.NoError(t, err)
	noSecretAccount, err := accountStore.Create("bad user", []byte("password"))
	require.NoError(t, err)

	totpCache := mock.NewTOTPCache(noSecretAccount.ID)
	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	require.NoError(t, totpCache.CacheTOTPSecret(account.ID, []byte(totpSecret)))
	require.NoError(t, totpCache.CacheTOTPSecret(failSetAccount.ID, []byte(totpSecret)))

	t.Run("no code", func(t *testing.T) {
		setErr := services.TOTPSetter(nil, nil, nil, 0, "")
		assert.Error(t, setErr)

		var v services.FieldErrors
		if errors.As(setErr, &v) {
			assert.Equal(t, v[0].Field, "otp")
			assert.Equal(t, v[0].Message, services.ErrInvalidOrExpired)
		} else {
			t.Fatalf("unexpected error type: %T", v)
		}
	})

	t.Run("no account", func(t *testing.T) {
		setErr := services.TOTPSetter(accountStore, nil, nil, 0, "")
		assert.Error(t, setErr)
	})

	t.Run("no secret in cache", func(t *testing.T) {
		setErr := services.TOTPSetter(accountStore, totpCache, nil, noSecretAccount.ID, "xxx")
		assert.Error(t, setErr)
	})

	t.Run("encrypt failure", func(t *testing.T) {
		code, generateErr := totp.GenerateCode(totpSecret, time.Now())
		require.NoError(t, generateErr)
		// Invalid key length will cause error
		setErr := services.TOTPSetter(accountStore, totpCache, &app.Config{DBEncryptionKey: []byte("XXX")}, account.ID, code)
		assert.Error(t, setErr)
	})

	t.Run("set failure", func(t *testing.T) {
		code, generateErr := totp.GenerateCode(totpSecret, time.Now())
		require.NoError(t, generateErr)
		// Invalid key length will cause error
		setErr := services.TOTPSetter(accountStore, totpCache, &app.Config{DBEncryptionKey: []byte("XXXXXXXXXXXXXXXX")}, failSetAccount.ID, code)
		assert.Error(t, setErr)
	})

	t.Run("happy", func(t *testing.T) {
		code, generateErr := totp.GenerateCode(totpSecret, time.Now())
		require.NoError(t, generateErr)
		setErr := services.TOTPSetter(accountStore, totpCache, &app.Config{DBEncryptionKey: []byte("XXXXXXXXXXXXXXXX")}, account.ID, code)
		assert.NoError(t, setErr)

		cachedSecret, checkErr := totpCache.LoadTOTPSecret(account.ID)
		assert.NoError(t, checkErr)
		assert.Nil(t, cachedSecret)

		t.Run("set unaffected", func(t *testing.T) {
			// re-cache the secret - we want this to try to set secret again
			require.NoError(t, totpCache.CacheTOTPSecret(account.ID, []byte(totpSecret)))
			// the mock account store is coded internally to return "unaffected" from SetTOTPSecret
			// if it receives the secret already set on the account found from lookup.
			// So if we try to set the same secret again we should get an error.
			setErr = services.TOTPSetter(accountStore, totpCache, &app.Config{DBEncryptionKey: []byte("XXXXXXXXXXXXXXXX")}, account.ID, code)
			assert.Error(t, setErr)

			cachedSecret, checkErr = totpCache.LoadTOTPSecret(account.ID)
			assert.NoError(t, checkErr)
			assert.NotNil(t, cachedSecret)

			assert.NoError(t, totpCache.RemoveTOTPSecret(account.ID))
		})
	})
}
