package data_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/keratin/authn-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore(t *testing.T) {
	testers := []func(*testing.T, data.RefreshTokenStore){
		testRefreshTokenFind,
		testRefreshTokenTouch,
		testRefreshTokenFindAll,
		testRefreshTokenCreate,
		testRefreshTokenRevoke,
	}

	t.Run("Mock", func(t *testing.T) {
		for _, tester := range testers {
			store := mock.NewRefreshTokenStore()
			tester(t, store)
		}
	})

	t.Run("Redis", func(t *testing.T) {
		client, err := redis.TestDB()
		require.NoError(t, err)
		store := &redis.RefreshTokenStore{Client: client, TTL: time.Second}
		for _, tester := range testers {
			tester(t, store)
			store.FlushDb()
		}
	})

	t.Run("Sqlite3", func(t *testing.T) {
		for _, tester := range testers {
			db, err := sqlite3.TestDB()
			require.NoError(t, err)
			store := &sqlite3.RefreshTokenStore{db, time.Second}
			tester(t, store)
			store.Close()
		}
	})
}

// TODO: find way to test that expired tokens are not found
func testRefreshTokenFind(t *testing.T, store data.RefreshTokenStore) {
	// finding nothing
	id, err := store.Find(models.RefreshToken("a1b2c3"))
	assert.Empty(t, id)
	assert.NoError(t, err)

	// finding something
	id2 := 123
	token, err := store.Create(id2)
	require.NoError(t, err)
	found, err := store.Find(token)
	if assert.NoError(t, err) {
		assert.Equal(t, found, id2)
	}
}

// TODO: find way to test for not touching expired tokens
func testRefreshTokenTouch(t *testing.T, store data.RefreshTokenStore) {
	err := store.Touch(models.RefreshToken("a1b2c3"), 123)
	assert.NoError(t, err)
}

// TODO: find way to test for not finding expired tokens
func testRefreshTokenFindAll(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	// finding nothing
	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Len(t, tokens, 0)

	token, err := store.Create(id)
	require.NoError(t, err)

	// finding something
	tokens2, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens2)
}

func testRefreshTokenCreate(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	token, err := store.Create(id)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens)
}

func testRefreshTokenRevoke(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	err := store.Revoke(models.RefreshToken("a1b2c3"))
	assert.NoError(t, err)

	token, err := store.Create(id)
	require.NoError(t, err)

	found, err := store.Find(token)
	if assert.NoError(t, err) {
		assert.Equal(t, found, id)
	}

	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens)

	err = store.Revoke(token)
	assert.NoError(t, err)

	found2, err := store.Find(token)
	assert.Empty(t, found2)
	assert.NoError(t, err)

	tokens2, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Len(t, tokens2, 0)
}
