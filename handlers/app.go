package handlers

import (
	"github.com/go-redis/redis"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"

	dataRedis "github.com/keratin/authn/data/redis"
)

type Pinger func() bool

type App struct {
	DbCheck           Pinger
	RedisCheck        Pinger
	Config            *config.Config
	AccountStore      data.AccountStore
	RefreshTokenStore data.RefreshTokenStore
}

func NewApp() (*App, error) {
	cfg := config.ReadEnv()

	db, accountStore, err := data.NewDB(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	opts, err := redis.ParseURL(cfg.RedisURL.String())
	if err != nil {
		return nil, err
	}
	redis := redis.NewClient(opts)

	tokenStore := &dataRedis.RefreshTokenStore{
		Client: redis,
		TTL:    cfg.RefreshTokenTTL,
	}

	return &App{
		DbCheck:           func() bool { return db.Ping() == nil },
		RedisCheck:        func() bool { return redis.Ping().Err() == nil },
		Config:            cfg,
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
	}, nil
}
