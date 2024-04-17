package mysql_test

import (
	"database/sql"
	"testing"

	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/data/mysql"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestAccountStore(t *testing.T) {
	db, err := mysql.TestDB()
	require.NoError(t, err)
	var store data.AccountStore = &mysql.AccountStore{db}
	for _, tester := range testers.AccountStoreTesters {
		db.MustExec("TRUNCATE accounts")
		db.MustExec("TRUNCATE oauth_accounts")
		tester(t, store)
	}

	t.Run("handle oauth email with null value", func(t *testing.T) {
		db := store.(interface {
			Exec(query string, args ...interface{}) (sql.Result, error)
		})

		account, err := store.Create("migrated-user", []byte("old"))
		require.NoError(t, err)

		err = store.AddOauthAccount(account.ID, "provider", "provider_id", "", "token")
		require.NoError(t, err)

		result, err := db.Exec("UPDATE oauth_accounts SET email = NULL WHERE account_id = ?", account.ID)
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)

		require.Equal(t, int64(1), rowsAffected)

		oAccounts, err := store.GetOauthAccounts(account.ID)
		require.NoError(t, err)

		require.Len(t, oAccounts, 1)
		require.True(t, oAccounts[0].Email == nil)
		require.Equal(t, oAccounts[0].GetEmail(), "")
	})
}
