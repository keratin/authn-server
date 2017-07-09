package sqlite3_test

import (
	"testing"

	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStore() sqlite3.AccountStore {
	db, err := tempDB()
	if err != nil {
		panic(err)
	}
	return sqlite3.AccountStore{db}
}

func TestCreate(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	assert.NoError(t, err)
	assert.NotEqual(t, 0, account.Id)
	assert.Equal(t, "authn@keratin.tech", account.Username)
	assert.NotEmpty(t, account.PasswordChangedAt)
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

func TestFindByUsername(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.FindByUsername("authn@keratin.tech")
	assert.NoError(t, err)
	assert.Nil(t, account)

	_, err = store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)

	account, err = store.FindByUsername("authn@keratin.tech")
	assert.NoError(t, err)
	assert.NotNil(t, account)
}

func TestLockAndUnlock(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.Locked)

	err = store.Lock(account.Id)
	require.NoError(t, err)

	account, err = store.Find(account.Id)
	require.NoError(t, err)
	assert.True(t, account.Locked)

	err = store.Unlock(account.Id)
	require.NoError(t, err)

	account, err = store.Find(account.Id)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	assert.False(t, account.Locked)
}

func TestArchive(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.Empty(t, account.DeletedAt)

	err = store.Archive(account.Id)
	require.NoError(t, err)

	account, err = store.Find(account.Id)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	assert.Empty(t, account.Username)
	assert.Empty(t, account.Password)
	assert.NotEmpty(t, account.DeletedAt)
}

func TestRequireNewPassword(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.RequireNewPassword)

	err = store.RequireNewPassword(account.Id)
	require.NoError(t, err)

	account, err = store.Find(account.Id)
	require.NoError(t, err)
	assert.True(t, account.RequireNewPassword)
}

func TestSetPassword(t *testing.T) {
	store := newStore()
	defer store.Close()

	account, err := store.Create("authn@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = store.RequireNewPassword(account.Id)
	require.NoError(t, err)

	err = store.SetPassword(account.Id, []byte("new"))
	require.NoError(t, err)

	account, err = store.Find(account.Id)
	require.NoError(t, err)
	assert.Equal(t, []byte("new"), account.Password)
	assert.False(t, account.RequireNewPassword)
}
