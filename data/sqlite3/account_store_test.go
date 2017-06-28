package sqlite3_test

import (
	"testing"

	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	assert.NoError(t, err)
	assert.NotEqual(t, 0, account.Id)
	assert.Equal(t, "authn@keratin.tech", account.Username)
	assert.NotEmpty(t, account.CreatedAt)
	assert.NotEmpty(t, account.UpdatedAt)

	account, err = store.Create("authn@keratin.tech", []byte("password"))
	if account != nil {
		assert.NotEqual(t, nil, account)
	}
	if !data.IsUniquenessError(err) {
		t.Errorf("expected uniqueness error, got %T %v", err, err)
	}
}
