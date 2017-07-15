package data_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountStore(t *testing.T) {
	testers := []func(*testing.T, data.AccountStore){
		testCreate,
		testFindByUsername,
		testLockAndUnlock,
		testArchive,
		testRequireNewPassword,
		testSetPassword,
	}

	t.Run("Mock", func(t *testing.T) {
		for _, tester := range testers {
			store := mock.NewAccountStore()
			tester(t, store)
		}
	})

	t.Run("Sqlite3", func(t *testing.T) {
		for _, tester := range testers {
			db, err := tempSqliteDB()
			if err != nil {
				panic(err)
			}
			store := sqlite3.AccountStore{db}
			tester(t, &store)
			store.Close()
		}
	})
}

func tempSqliteDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	err = sqlite3.MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func testCreate(t *testing.T, store data.AccountStore) {
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

func testFindByUsername(t *testing.T, store data.AccountStore) {
	account, err := store.FindByUsername("authn@keratin.tech")
	assert.NoError(t, err)
	assert.Nil(t, account)

	_, err = store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)

	account, err = store.FindByUsername("authn@keratin.tech")
	assert.NoError(t, err)
	assert.NotNil(t, account)
}

func testLockAndUnlock(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.Locked)

	err = store.Lock(account.Id)
	require.NoError(t, err)

	after, err := store.Find(account.Id)
	require.NoError(t, err)
	assert.True(t, after.Locked)

	err = store.Unlock(account.Id)
	require.NoError(t, err)

	after2, err := store.Find(account.Id)
	require.NoError(t, err)
	require.NotEmpty(t, after2)
	assert.False(t, after2.Locked)
}

func testArchive(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.Empty(t, account.DeletedAt)

	err = store.Archive(account.Id)
	require.NoError(t, err)

	after, err := store.Find(account.Id)
	require.NoError(t, err)
	require.NotEmpty(t, after)
	assert.Empty(t, after.Username)
	assert.Empty(t, after.Password)
	assert.NotEmpty(t, after.DeletedAt)
}

func testRequireNewPassword(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.RequireNewPassword)

	err = store.RequireNewPassword(account.Id)
	require.NoError(t, err)

	after, err := store.Find(account.Id)
	require.NoError(t, err)
	assert.True(t, after.RequireNewPassword)
}

func testSetPassword(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = store.RequireNewPassword(account.Id)
	require.NoError(t, err)

	err = store.SetPassword(account.Id, []byte("new"))
	require.NoError(t, err)

	after, err := store.Find(account.Id)
	require.NoError(t, err)
	assert.Equal(t, []byte("new"), after.Password)
	assert.False(t, after.RequireNewPassword)
	assert.NotEqual(t, account.PasswordChangedAt, after.PasswordChangedAt)
}
