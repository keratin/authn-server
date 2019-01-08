package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsVerifierSuccess(t *testing.T) {
	username := "myname"
	password := "mysecret"
	bcrypted := []byte("$2a$04$lzQPXlov4RFLxps1uUGq4e4wmVjLYz3WrqQw4bSdfIiJRyo3/fk3C")

	cfg := app.Config{BcryptCost: 4}
	store := mock.NewAccountStore()
	store.Create(username, bcrypted)

	acc, err := services.CredentialsVerifier(store, &cfg, username, password)
	require.NoError(t, err)
	assert.NotEqual(t, 0, acc.ID)
	assert.Equal(t, username, acc.Username)
}

func TestCredentialsVerifierFailure(t *testing.T) {
	password := "mysecret"
	bcrypted := []byte("$2a$04$lzQPXlov4RFLxps1uUGq4e4wmVjLYz3WrqQw4bSdfIiJRyo3/fk3C")

	cfg := app.Config{BcryptCost: 4}
	store := mock.NewAccountStore()
	store.Create("known", bcrypted)
	acc, _ := store.Create("locked", bcrypted)
	store.Lock(acc.ID)
	acc, _ = store.Create("expired", bcrypted)
	store.RequireNewPassword(acc.ID)

	testCases := []struct {
		username string
		password string
		errors   services.FieldErrors
	}{
		{"", "", services.FieldErrors{{"credentials", "FAILED"}}},
		{"unknown", "", services.FieldErrors{{"credentials", "FAILED"}}},
		{"unknown", "unknown", services.FieldErrors{{"credentials", "FAILED"}}},
		{"known", "unknown", services.FieldErrors{{"credentials", "FAILED"}}},
		{"unknown", password, services.FieldErrors{{"credentials", "FAILED"}}},
		{"locked", password, services.FieldErrors{{"account", "LOCKED"}}},
		{"expired", password, services.FieldErrors{{"credentials", "EXPIRED"}}},
	}

	for _, tc := range testCases {
		_, errs := services.CredentialsVerifier(store, &cfg, tc.username, tc.password)
		assert.Equal(t, tc.errors, errs)
	}
}
