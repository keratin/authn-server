package redis

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
)

func New(url *url.URL) (*redis.Client, error) {
	return newFromString(url.String())
}

func NewSentinel(redisSentinelMaster string, redisSentinelNodes string, redisSentinelPassword string) (*redis.Client, error) {
	sentinelAddressSlice := strings.Split(redisSentinelNodes, ",")
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    redisSentinelMaster,
		SentinelAddrs: sentinelAddressSlice,
		Password:      redisSentinelPassword,
	}), nil
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
