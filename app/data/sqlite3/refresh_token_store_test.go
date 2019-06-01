package sqlite3_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/sqlite3"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore(t *testing.T) {
	for _, tester := range testers.RefreshTokenStoreTesters {
		db, err := sqlite3.TestDB()
		require.NoError(t, err)
		store := &sqlite3.RefreshTokenStore{db, time.Second}
		tester(t, store)
		db.Close()
	}
}
