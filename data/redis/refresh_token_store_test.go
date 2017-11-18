package redis_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/data/testers"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore(t *testing.T) {
	client, err := redis.TestDB()
	require.NoError(t, err)
	store := &redis.RefreshTokenStore{Client: client, TTL: time.Second}
	for _, tester := range testers.RefreshTokenStoreTesters {
		tester(t, store)
		store.FlushDb()
	}
}
