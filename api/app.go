package api

import (
	"os"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/keratin/authn-server/data/mock"
	dataRedis "github.com/keratin/authn-server/data/redis"
)

type pinger func() bool

type App struct {
	DbCheck           pinger
	RedisCheck        pinger
	Config            *config.Config
	AccountStore      data.AccountStore
	RefreshTokenStore data.RefreshTokenStore
	KeyStore          data.KeyStore
	Actives           data.Actives
	Reporter          ops.ErrorReporter
}

func NewApp() (*App, error) {
	cfg := config.ReadEnv()

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)

	var reporter ops.ErrorReporter
	if cfg.SentryDSN != "" {
		c, err := raven.New(cfg.SentryDSN)
		if err != nil {
			return nil, errors.Wrap(err, "raven.New")
		}
		reporter = &ops.SentryReporter{Client: c}
	} else {
		reporter = &ops.LogReporter{}
	}

	db, accountStore, err := data.NewDB(cfg.DatabaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "data.NewDB")
	}

	redis, err := dataRedis.New(cfg.RedisURL)
	if err != nil {
		return nil, errors.Wrap(err, "redis.New")
	}

	tokenStore := &dataRedis.RefreshTokenStore{
		Client: redis,
		TTL:    cfg.RefreshTokenTTL,
	}

	var keyStore data.KeyStore
	if cfg.IdentitySigningKey == nil {
		keyStore, err = dataRedis.NewKeyStore(
			redis,
			reporter,
			cfg.AccessTokenTTL,
			time.Duration(500)*time.Millisecond,
			cfg.DBEncryptionKey,
		)
		if err != nil {
			return nil, errors.Wrap(err, "NewKeyStore")
		}
	} else {
		keyStore = mock.NewKeyStore(cfg.IdentitySigningKey)
	}

	actives := dataRedis.NewActives(
		redis,
		cfg.StatisticsTimeZone,
		cfg.DailyActivesRetention,
		cfg.WeeklyActivesRetention,
		5*12,
	)

	return &App{
		DbCheck:           func() bool { return db.Ping() == nil },
		RedisCheck:        func() bool { return redis.Ping().Err() == nil },
		Config:            cfg,
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
		KeyStore:          keyStore,
		Actives:           actives,
		Reporter:          reporter,
	}, nil
}
