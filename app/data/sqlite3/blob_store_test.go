package sqlite3_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/sqlite3"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestBlobStore(t *testing.T) {
	for _, tester := range testers.BlobStoreTesters {
		db, err := sqlite3.TestDB()
		require.NoError(t, err)
		store := &sqlite3.BlobStore{
			TTL:      time.Minute,
			LockTime: time.Minute,
			DB:       db,
		}
		tester(t, store)
		db.Close()
	}
}
