package api

import (
	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"

	dataRedis "github.com/keratin/authn-server/data/redis"
)

type pinger func() bool

type App struct {
	DbCheck           pinger
	RedisCheck        pinger
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
