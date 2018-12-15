package mysql_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mysql"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestAccountStore(t *testing.T) {
	db, err := mysql.TestDB()
	require.NoError(t, err)
	store := &mysql.AccountStore{db}
	for _, tester := range testers.AccountStoreTesters {
		db.MustExec("TRUNCATE accounts")
		db.MustExec("TRUNCATE oauth_accounts")
		tester(t, store)
	}
}
