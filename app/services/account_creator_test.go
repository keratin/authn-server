package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCreatorSuccess(t *testing.T) {
	store := mock.NewAccountStore()

	var testCases = []struct {
		config   app.Config
		username string
		password string
	}{
		{app.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "userName", "PASSword"},
		{app.Config{UsernameIsEmail: true}, "username@test.com", "PASSword"},
		{app.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "username@rightdomain.com", "PASSword"},
	}

	for _, tc := range testCases {
		acc, err := services.AccountCreator(store, &tc.config, tc.username, tc.password)
		require.NoError(t, err)
		assert.NotEqual(t, 0, acc.ID)
		assert.Equal(t, tc.username, acc.Username)
	}
}

var pw = []byte("$2a$04$ZOBA8E3nT68/ArE6NDnzfezGWEgM6YrE17PrOtSjT5.U/ZGoxyh7e")

func TestAccountCreatorFailure(t *testing.T) {
	store := mock.NewAccountStore()
	store.Create("existing@test.com", pw)

	var testCases = []struct {
		config   app.Config
		username string
		password string
		errors   services.FieldErrors
	}{
		// username validations
		{app.Config{}, "", "PASSword", services.FieldErrors{{"username", "MISSING"}}},
		{app.Config{}, "  ", "PASSword", services.FieldErrors{{"username", "MISSING"}}},
		{app.Config{}, "existing@test.com", "PASSword", services.FieldErrors{{"username", "TAKEN"}}},
		{app.Config{UsernameIsEmail: true}, "notanemail", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{app.Config{UsernameIsEmail: true}, "@wrong.com", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{app.Config{UsernameIsEmail: true}, "wrong@wrong", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{app.Config{UsernameIsEmail: true}, "wrong@wrong.", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{app.Config{UsernameIsEmail: true, UsernameDomains: []string{"rightdomain.com"}}, "email@wrongdomain.com", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		{app.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "short", "PASSword", services.FieldErrors{{"username", "FORMAT_INVALID"}}},
		// password validations
		{app.Config{}, "username", "", services.FieldErrors{{"password", "MISSING"}}},
		{app.Config{PasswordMinComplexity: 2}, "username", "qwerty", services.FieldErrors{{"password", "INSECURE"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			acc, err := services.AccountCreator(store, &tc.config, tc.username, tc.password)
			if assert.Equal(t, tc.errors, err) {
				assert.Empty(t, acc)
			}
		})
	}
}
