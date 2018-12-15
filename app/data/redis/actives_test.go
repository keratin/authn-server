package redis_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/redis"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestActives(t *testing.T) {
	client, err := redis.TestDB()
	require.NoError(t, err)
	rStore := redis.NewActives(client, time.UTC, 365, 52, 12)
	for _, tester := range testers.ActivesTesters {
		client.FlushDB()
		tester(t, rStore)
	}
}
