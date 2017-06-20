package handlers

import (
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type Pinger func() bool

type App struct {
	DbCheck           Pinger
	RedisCheck        Pinger
	Config            *config.Config
	AccountStore      data.AccountStore
	RefreshTokenStore data.RefreshTokenStore
}
