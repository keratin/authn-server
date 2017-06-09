package redis_test

import (
	"reflect"
	"testing"
	"time"

	goredis "github.com/go-redis/redis"
	"github.com/keratin/authn/data"
	"github.com/keratin/authn/data/redis"
)

var oneDay = time.Duration(8766) * time.Hour
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
	store := redis.RefreshTokenStore{Client: client, TTL: oneDay}

	token := data.RefreshToken("a1b2c3")
	key := "s:t.\xa1\xb2\xc3"
	id := 123

	expectNoId(t, func() (int, error) {
		return store.Find(token)
	})

	// insert into redis
	err := client.Set(key, id, 0).Err()
	if err != nil {
		t.Fatal(err)
	}

	expectId(id, t, func() (int, error) {
		return store.Find(token)
	})
}

func TestRefreshTokenTouch(t *testing.T) {
}

func TestRefreshTokenFindAll(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: oneDay}

	id := 123

	expectNoTokens(t, func() ([]data.RefreshToken, error) {
		return store.FindAll(id)
	})

	// insert
	token, err := store.Create(id)
	if err != nil {
		t.Fatal(err)
	}

	expectTokens([]data.RefreshToken{token}, t, func() ([]data.RefreshToken, error) {
		return store.FindAll(id)
	})
}

func TestRefreshTokenCreate(t *testing.T) {
}

func TestRefreshTokenRevoke(t *testing.T) {
	client := newStore()
	defer client.FlushDb()
	store := redis.RefreshTokenStore{Client: client, TTL: oneDay}

	id := 123

	err := store.Revoke(data.RefreshToken("a1b2c3"))
	if err != nil {
		t.Error(err)
	}

	token, err := store.Create(id)
	if err != nil {
		t.Fatal(err)
	}

	expectId(id, t, func() (int, error) {
		return store.Find(token)
	})
	expectTokens([]data.RefreshToken{token}, t, func() ([]data.RefreshToken, error) {
		return store.FindAll(id)
	})

	err = store.Revoke(token)
	if err != nil {
		t.Error(err)
	}

	expectNoId(t, func() (int, error) {
		return store.Find(token)
	})
	expectNoTokens(t, func() ([]data.RefreshToken, error) {
		return store.FindAll(id)
	})
}

type ider func() (int, error)

func expectNoId(t *testing.T, fn ider) {
	id, err := fn()
	if id != 0 {
		t.Error("expected empty value, got %v", id)
	}
	if err != nil {
		t.Error(err)
	}
}

func expectId(expected int, t *testing.T, fn ider) {
	id, err := fn()
	if err != nil {
		t.Error(err)
	} else if expected != id {
		t.Errorf("expected: %v, got: %v", expected, id)
	}
}

type tokenser func() ([]data.RefreshToken, error)

func expectNoTokens(t *testing.T, fn tokenser) {
	tokens, err := fn()
	if err != nil {
		t.Error(err)
	}
	if len(tokens) > 0 {
		t.Errorf("expected no tokens, got %v", tokens)
	}
}

func expectTokens(expected []data.RefreshToken, t *testing.T, fn tokenser) {
	tokens, err := fn()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, tokens) {
		t.Errorf("expected %v, got %v", expected, tokens)
	}
}
