package testers

import (
	"database/sql"
	"testing"

	"github.com/keratin/authn-server/app/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var AccountStoreTesters = []func(*testing.T, data.AccountStore){
	testCreate,
	testFindByUsername,
	testLockAndUnlock,
	testArchive,
	testArchiveWithOauth,
	testRequireNewPassword,
	testSetPassword,
	testSetAndDeleteTOTP,
	testUpdateUsername,
	testAddOauthAccount,
	testFindByOauthAccount,
	testSetLastLogin,
}

type hasStats interface {
	Stats() sql.DBStats
}

func getOpenConnectionCount(store data.AccountStore) int {
	if st, ok := store.(hasStats); ok {
		return st.Stats().OpenConnections
	} else {
		return 1
	}
}

func testCreate(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	assert.NotEqual(t, 0, account.ID)
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

	account, err = store.Create("AUTHN@KERATIN.TECH", []byte("password"))
	if account != nil {
		assert.NotEqual(t, nil, account)
	}
	if !data.IsUniquenessError(err) {
		t.Errorf("expected uniqueness error, got %T %v", err, err)
	}

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
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

	account, err = store.FindByUsername("AUTHN@KERATIN.TECH")
	assert.NoError(t, err)
	assert.NotNil(t, account)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testLockAndUnlock(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.Locked)

	ok, err := store.Lock(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.True(t, after.Locked)

	ok, err = store.Unlock(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	after2, err := store.Find(account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, after2)
	assert.False(t, after2.Locked)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testArchive(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.Empty(t, account.DeletedAt)

	ok, err := store.Archive(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, after)
	assert.Empty(t, after.Username)
	assert.Empty(t, after.Password)
	assert.NotEmpty(t, after.DeletedAt)

	account2, err := store.Create("authn@keratin.tech", []byte("password"))
	if assert.NoError(t, err) {
		ok, err = store.Archive(account2.ID)
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testArchiveWithOauth(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	err = store.AddOauthAccount(account.ID, "PROVIDER", "PROVIDERID", "email", "token")
	require.NoError(t, err)

	ok, err := store.Archive(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	found, err := store.FindByOauthAccount("PROVIDER", "PROVIDERID")
	require.NoError(t, err)
	assert.Empty(t, found)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testRequireNewPassword(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	require.False(t, account.RequireNewPassword)

	ok, err := store.RequireNewPassword(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.True(t, after.RequireNewPassword)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testSetPassword(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("old"))
	require.NoError(t, err)
	ok, err := store.RequireNewPassword(account.ID)
	require.True(t, ok)
	require.NoError(t, err)

	ok, err = store.SetPassword(account.ID, []byte("new"))
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.Equal(t, []byte("new"), after.Password)
	assert.False(t, after.RequireNewPassword)
	assert.NotEqual(t, account.PasswordChangedAt, after.PasswordChangedAt)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testSetAndDeleteTOTP(t *testing.T, store data.AccountStore) {
	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	assert.False(t, account.TOTPEnabled())
	assert.False(t, account.TOTPSecret.Valid)

	//Check set
	ok, err := store.SetTOTPSecret(account.ID, []byte("secret"))
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.Equal(t, "secret", after.TOTPSecret.String)
	assert.True(t, after.TOTPEnabled())
	assert.True(t, after.TOTPSecret.Valid)

	//Check delete
	ok, err = store.DeleteTOTPSecret(account.ID)
	assert.True(t, ok)
	require.NoError(t, err)

	after, err = store.Find(account.ID)
	require.NoError(t, err)
	assert.False(t, after.TOTPEnabled())
	assert.False(t, after.TOTPSecret.Valid)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testUpdateUsername(t *testing.T, store data.AccountStore) {
	other, err := store.Create("other", []byte("other"))
	require.NoError(t, err)

	account, err := store.Create("old", []byte("old"))
	require.NoError(t, err)

	ok, err := store.UpdateUsername(account.ID, "new")
	assert.True(t, ok)
	require.NoError(t, err)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.Equal(t, "new", after.Username)

	ok, err = store.UpdateUsername(account.ID, other.Username)
	assert.False(t, ok)
	if err == nil || !data.IsUniquenessError(err) {
		t.Errorf("expected uniqueness error, got %T %v", err, err)
	}

	// "changing" to existing username
	_, err = store.UpdateUsername(account.ID, "new")
	require.NoError(t, err)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testAddOauthAccount(t *testing.T, store data.AccountStore) {
	found, err := store.GetOauthAccounts(1)
	require.NoError(t, err)
	assert.Len(t, found, 0)

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	assert.NoError(t, err)
	err = store.AddOauthAccount(account.ID, "OAUTHPROVIDER", "PROVIDERID", "email", "TOKEN")
	assert.NoError(t, err)

	found, err = store.GetOauthAccounts(account.ID)
	require.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, account.ID, found[0].AccountID)
	assert.Equal(t, "OAUTHPROVIDER", found[0].Provider)
	assert.Equal(t, "PROVIDERID", found[0].ProviderID)
	assert.Equal(t, "TOKEN", found[0].AccessToken)
	assert.NotEmpty(t, found[0].CreatedAt)
	assert.NotEmpty(t, found[0].UpdatedAt)

	err = store.AddOauthAccount(account.ID, "OAUTHPROVIDER", "PROVIDERID2", "email", "TOKEN")
	if err == nil || !data.IsUniquenessError(err) {
		t.Errorf("expected uniqueness error, got %T %v", err, err)
	}

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testFindByOauthAccount(t *testing.T, store data.AccountStore) {
	found, err := store.FindByOauthAccount("unknown", "unknown")
	assert.NoError(t, err)
	assert.Nil(t, found)

	account, err := store.Create("authn@keratin.tech", []byte("password"))
	require.NoError(t, err)
	err = store.AddOauthAccount(account.ID, "OAUTHPROVIDER", "PROVIDERID", "email", "TOKEN")
	require.NoError(t, err)

	found, err = store.FindByOauthAccount("unknown", "PROVIDERID")
	assert.NoError(t, err)
	assert.Nil(t, found)

	found, err = store.FindByOauthAccount("OAUTHPROVIDER", "unknown")
	assert.NoError(t, err)
	assert.Nil(t, found)

	found, err = store.FindByOauthAccount("OAUTHPROVIDER", "PROVIDERID")
	assert.NoError(t, err)
	assert.Equal(t, account.ID, found.ID)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}

func testSetLastLogin(t *testing.T, store data.AccountStore) {
	account, err := store.Create("old", []byte("old"))
	require.NoError(t, err)

	rowsIsAffected, err := store.SetLastLogin(account.ID)
	require.NoError(t, err)
	require.Equal(t, true, rowsIsAffected)

	after, err := store.Find(account.ID)
	require.NoError(t, err)
	assert.NotEqual(t, nil, after.LastLoginAt)

	// Assert that db connections are released to pool
	assert.Equal(t, 1, getOpenConnectionCount(store))
}
