package sqlite3_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/sqlite3"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestAccountStore(t *testing.T) {
	for _, tester := range testers.AccountStoreTesters {
		db, err := sqlite3.TestDB()
		require.NoError(t, err)
		store := &sqlite3.AccountStore{db}
		tester(t, store)
		db.Close()
	}
}
