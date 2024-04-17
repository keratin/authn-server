package postgres_test

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/data/postgres"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func newTestDB() (*sqlx.DB, error) {
	str, ok := os.LookupEnv("TEST_POSTGRES_URL")
	if !ok {
		return nil, fmt.Errorf("set TEST_POSTGRES_URL for PostgreSQL tests")
	}

	dbURL, err := url.Parse(str)
	if err != nil {
		return nil, err
	}

	db, err := postgres.NewDB(dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "NewDB")
	}

	if err := postgres.MigrateDB(db); err != nil {
		return nil, errors.Wrap(err, "MigrateDB")
	}

	return db, nil
}
func TestAccountStore(t *testing.T) {
	db, err := newTestDB()
	require.NoError(t, err)
	var store data.AccountStore = &postgres.AccountStore{db}
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

		result, err := db.Exec("UPDATE oauth_accounts SET email = NULL WHERE account_id = $1", account.ID)
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
