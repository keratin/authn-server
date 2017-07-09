package services_test

import (
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCreatorSuccess(t *testing.T) {
	store := mock.NewAccountStore()

	var testCases = []struct {
		config   config.Config
		username string
		password string
	}{
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "userName", "PASSword"},
		{config.Config{UsernameIsEmail: true}, "username@test.com", "PASSword"},
		{config.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "username@rightdomain.com", "PASSword"},
	}

	for _, tc := range testCases {
		acc, err := services.AccountCreator(store, &tc.config, tc.username, tc.password)
		require.NoError(t, err)
		assert.NotEqual(t, 0, acc.Id)
		assert.Equal(t, tc.username, acc.Username)
	}
}

var pw = []byte("$2a$04$ZOBA8E3nT68/ArE6NDnzfezGWEgM6YrE17PrOtSjT5.U/ZGoxyh7e")

func TestAccountCreatorFailure(t *testing.T) {
	store := mock.NewAccountStore()
	store.Create("existing@test.com", pw)

	var testCases = []struct {
		config   config.Config
		username string
		password string
		errors   services.FieldErrors
	}{
		// username validations
		{config.Config{}, "", "PASSword", services.FieldErrors{{"username", "MISSING"}}},
		{config.Config{}, "  ", "PASSword", services.FieldErrors{{"username", "MISSING"}}},
		{config.Config{}, "existing@test.com", "PASSword", services.FieldErrors{{"username", "TAKEN"}}},
		{config.Config{UsernameIsEmail: true}, "notanemail", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "email@wrongdomain.com", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "short", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		// password validations
		{config.Config{}, "username", "", services.FieldErrors{{"password", "MISSING"}}},
		{config.Config{PasswordMinComplexity: 2}, "username", "qwerty", services.FieldErrors{{"password", "INSECURE"}}},
	}

	for _, tc := range testCases {
		acc, err := services.AccountCreator(store, &tc.config, tc.username, tc.password)
		assert.Empty(t, acc)
		assert.Equal(t, tc.errors, err)
	}
}
