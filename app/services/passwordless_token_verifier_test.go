package services_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/keratin/authn-server/ops"
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordlessTokenVerifier(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		AuthNURL:                    &url.URL{Scheme: "http", Host: "authn.example.com"},
		BcryptCost:                  4,
		PasswordMinComplexity:       1,
		PasswordlessTokenSigningKey: []byte("reset-a-reno"),
	}

	newToken := func(id int) string {
		claims, err := passwordless.New(cfg, id)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		return token
	}

	invoke := func(token string) error {
		_, err := services.PasswordlessTokenVerifier(accountStore, &ops.LogReporter{FieldLogger: logrus.New()}, cfg, token, "")
		return err
	}

	t.Run("when token is valid", func(t *testing.T) {
		account, err := accountStore.Create("valid@keratin.tech", []byte("old"))
		require.NoError(t, err)

		token := newToken(account.ID)

		err = invoke(token)
		require.NoError(t, err)
	})

	t.Run("when token is invalid", func(t *testing.T) {
		// nolint: gosec
		token := "not.valid.jwt"

		err := invoke(token)
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	t.Run("on an archived account", func(t *testing.T) {
		archived, err := accountStore.Create("archived@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Archive(archived.ID)
		require.NoError(t, err)

		token := newToken(archived.ID)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("on a locked account", func(t *testing.T) {
		locked, err := accountStore.Create("locked@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Lock(locked.ID)
		require.NoError(t, err)

		token := newToken(locked.ID)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("when account has logged in again", func(t *testing.T) {
		account, err := accountStore.Create("account@keratin.tech", []byte("old"))
		require.NoError(t, err)

		token := newToken(account.ID)

		_, err = accountStore.SetLastLogin(account.ID)
		require.NoError(t, err)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		token := newToken(0)
		err := invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})
}

func TestPasswordlessTokenVerifierWithOTP(t *testing.T) {
	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	totpSecretEnc := []byte("cli6azfL5i7PAnh8U/w3Zbglsm3XcdaGODy+Ga5QqT02c9hotDAR1Y28--3UihzsJhw/+EU3R6--qUw9L8DwN5XPVfOStshKzA==")

	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		AuthNURL:                    &url.URL{Scheme: "http", Host: "authn.example.com"},
		BcryptCost:                  4,
		DBEncryptionKey:             []byte("DLz2TNDRdWWA5w8YNeCJ7uzcS4WDzQmB"),
		PasswordMinComplexity:       1,
		PasswordlessTokenSigningKey: []byte("reset-a-reno"),
	}

	newToken := func(id int) string {
		claims, err := passwordless.New(cfg, id)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		return token
	}

	invoke := func(token string, totpCode string) error {
		_, err := services.PasswordlessTokenVerifier(accountStore, &ops.LogReporter{FieldLogger: logrus.New()}, cfg, token, totpCode)
		return err
	}

	t.Run("with good code", func(t *testing.T) {
		account, err := accountStore.Create("first@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.SetTOTPSecret(account.ID, totpSecretEnc)
		require.NoError(t, err)
		token := newToken(account.ID)

		code, err := totp.GenerateCode(totpSecret, time.Now())
		require.NoError(t, err)

		err = invoke(token, code)
		require.NoError(t, err)
	})

	t.Run("with bad code", func(t *testing.T) {
		account, err := accountStore.Create("second@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.SetTOTPSecret(account.ID, totpSecretEnc)
		require.NoError(t, err)
		token := newToken(account.ID)

		err = invoke(token, "12345")
		assert.Equal(t, services.FieldErrors{{"otp", "INVALID_OR_EXPIRED"}}, err)

		err = invoke(token, "")
		assert.Equal(t, services.FieldErrors{{"otp", "MISSING"}}, err)
	})
}
