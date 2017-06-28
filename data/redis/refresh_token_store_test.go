package redis_test

import (
	"testing"
	"time"

	goredis "github.com/go-redis/redis"
	"github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var refreshTTL = time.Second
var redisUrl = "redis://127.0.0.1:6379/12"

func newStore() *goredis.Client {
	opts, err := goredis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}
	return goredis.NewClient(opts)
}

func TestRefreshTokenFind(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: refreshTTL}

	token := models.RefreshToken("a1b2c3")
	key := "s:t.\xa1\xb2\xc3"
	id := 123

	expectNoId(t, func() (int, error) {
		return store.Find(token)
	})

	// insert into redis
	err := client.Set(key, id, 0).Err()
	require.NoError(t, err)

	expectId(id, t, func() (int, error) {
		return store.Find(token)
	})
}

func TestRefreshTokenTouch(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: refreshTTL}

	err := store.Touch(models.RefreshToken("a1b2c3"), 123)
	assert.NoError(t, err)
}

func TestRefreshTokenFindAll(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: refreshTTL}

	id := 123

	expectNoTokens(t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})

	// insert
	token, err := store.Create(id)
	require.NoError(t, err)

	expectTokens([]models.RefreshToken{token}, t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})
}

func TestRefreshTokenCreate(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: refreshTTL}

	id := 123

	expectNoTokens(t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})

	token, err := store.Create(id)
	assert.NoError(t, err)
	assert.Len(t, token, 32)

	expectTokens([]models.RefreshToken{token}, t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})
}

func TestRefreshTokenRevoke(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: refreshTTL}

	id := 123

	err := store.Revoke(models.RefreshToken("a1b2c3"))
	assert.NoError(t, err)

	token, err := store.Create(id)
	require.NoError(t, err)

	expectId(id, t, func() (int, error) {
		return store.Find(token)
	})
	expectTokens([]models.RefreshToken{token}, t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})

	err = store.Revoke(token)
	assert.NoError(t, err)

	expectNoId(t, func() (int, error) {
		return store.Find(token)
	})
	expectNoTokens(t, func() ([]models.RefreshToken, error) {
		return store.FindAll(id)
	})
}

type ider func() (int, error)

func expectNoId(t *testing.T, fn ider) {
	id, err := fn()
	assert.Empty(t, id)
	assert.NoError(t, err)
}

func expectId(expected int, t *testing.T, fn ider) {
	id, err := fn()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, id)
	}
}

type tokenser func() ([]models.RefreshToken, error)

func expectNoTokens(t *testing.T, fn tokenser) {
	tokens, err := fn()
	assert.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func expectTokens(expected []models.RefreshToken, t *testing.T, fn tokenser) {
	tokens, err := fn()
	assert.NoError(t, err)
	assert.Equal(t, expected, tokens)
}
