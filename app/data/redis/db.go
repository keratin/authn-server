package redis

import (
	"fmt"
	"net/url"
	"os"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

func New(url *url.URL) (*redis.Client, error) {
	opts, err := redis.ParseURL(url.String())
	if err != nil {
		return nil, errors.Wrap(err, "ParseURL")
	}
	return redis.NewClient(opts), nil
}

// TODO: move to _test
func TestDB() (*redis.Client, error) {
	str, ok := os.LookupEnv("TEST_REDIS_URL")
	if !ok {
		return nil, fmt.Errorf("set TEST_REDIS_URL for redis tests")
	}
	cfg, err := redis.ParseURL(str)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(cfg), nil
}
