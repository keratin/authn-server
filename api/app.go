package api

import (
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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

	db, err := data.NewDB(cfg.DatabaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "data.NewDB")
	}

	redis, err := dataRedis.New(cfg.RedisURL)
	if err != nil {
		return nil, errors.Wrap(err, "redis.New")
	}

	accountStore := data.NewAccountStore(db)
	if accountStore == nil {
		return nil, errors.Wrap(err, "NewAccountStore")
	}

	tokenStore := data.NewRefreshTokenStore(db, redis, reporter, cfg.RefreshTokenTTL)
	if tokenStore == nil {
		return nil, errors.Wrap(err, "NewRefreshTokenStore")
	}

	keyStore := data.NewRotatingKeyStore()
	if cfg.IdentitySigningKey == nil {
		m := data.NewKeyStoreRotater(
			data.NewEncryptedBlobStore(
				data.NewBlobStore(cfg.AccessTokenTTL, redis),
				cfg.DBEncryptionKey,
			),
			cfg.AccessTokenTTL,
		)
		err := m.Maintain(keyStore, reporter)
		if err != nil {
			return nil, errors.Wrap(err, "Maintain")
		}
	} else {
		keyStore.Rotate(cfg.IdentitySigningKey)
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
