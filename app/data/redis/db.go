package redis

import (
	"fmt"
	"net/url"
	"os"

	"github.com/go-redis/redis"
)

func New(url *url.URL) (*redis.Client, error) {
	return newFromString(url.String())
}

// TODO: move to _test
func TestDB() (*redis.Client, error) {
	str, ok := os.LookupEnv("TEST_REDIS_URL")
	if !ok {
		return nil, fmt.Errorf("set TEST_REDIS_URL for redis tests")
	}
	return newFromString(str)
}

func newFromString(url string) (*redis.Client, error) {
	cfg, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(cfg), nil
}
