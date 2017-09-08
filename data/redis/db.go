package redis

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

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
