package services_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// it's a "secret"
var bcrypted = []byte("$2a$10$W5AiL6r4XBrZHc3NEcMUC.xj52oYl6YQw6YpTP1OkjFLmWfOk7oqC")

func TestAccountImporter(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		BcryptCost: 4,
	}

	_, err := accountStore.Create("existing", []byte("secret"))
	require.NoError(t, err)

	testCases := []struct {
		username string
		password []byte
		locked   bool
		errors   *services.FieldErrors
	}{
		{"unlocked", bcrypted, false, nil},
		{"locked", bcrypted, true, nil},
		{"plaintext", []byte("secret"), false, nil},
		{"", bcrypted, false, &services.FieldErrors{{"username", services.ErrMissing}}},
		{"invalid", []byte(""), false, &services.FieldErrors{{"password", services.ErrMissing}}},
		{"existing", bcrypted, false, &services.FieldErrors{{"username", services.ErrTaken}}},
	}

	for _, tc := range testCases {
		account, errors := services.AccountImporter(accountStore, cfg, tc.username, string(tc.password), tc.locked)
		if tc.errors == nil {
			assert.Empty(t, errors)
			assert.NotEmpty(t, account)
			assert.Equal(t, tc.locked, account.Locked)
			assert.Equal(t, tc.username, account.Username)
			assert.NoError(t, bcrypt.CompareHashAndPassword(account.Password, []byte("secret")))
		} else {
			assert.Equal(t, *tc.errors, errors)
			assert.Empty(t, account)
		}
	}
}
