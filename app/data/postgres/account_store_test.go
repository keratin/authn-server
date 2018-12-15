package postgres_test

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
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
	store := &postgres.AccountStore{db}
	for _, tester := range testers.AccountStoreTesters {
		db.MustExec("TRUNCATE accounts")
		db.MustExec("TRUNCATE oauth_accounts")
		tester(t, store)
	}
}
