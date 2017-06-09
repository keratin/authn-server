package handlers

import (
	"database/sql"

	"github.com/go-redis/redis"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type App struct {
	Db                sql.DB
	Redis             *redis.Client
	Config            config.Config
	AccountStore      data.AccountStore
	RefreshTokenStore data.RefreshTokenStore
}
