package redis

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/go-redis/redis"
)

func TestDB() (*redis.Client, error) {
	if _, err := os.Stat("../.env"); !os.IsNotExist(err) {
		godotenv.Load("../.env")
	}
	if _, err := os.Stat("../../.env"); !os.IsNotExist(err) {
		godotenv.Load("../../.env")
	}

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
