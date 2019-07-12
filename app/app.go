package app

import (
	"os"

	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	dataRedis "github.com/keratin/authn-server/app/data/redis"
)

type pinger func() bool

type App struct {
	DB                *sqlx.DB
	DbCheck           pinger
	RedisCheck        pinger
	Config            *Config
	AccountStore      data.AccountStore
	RefreshTokenStore data.RefreshTokenStore
	KeyStore          data.KeyStore
	Actives           data.Actives
	Reporter          ops.ErrorReporter
	OauthProviders    map[string]oauth.Provider
}

func NewApp(cfg *Config) (*App, error) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)

	db, err := data.NewDB(cfg.DatabaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "data.NewDB")
	}

	var redis *redis.Client
	if cfg.RedisURL != nil {
		redis, err = dataRedis.New(cfg.RedisURL)
		if err != nil {
			return nil, errors.Wrap(err, "redis.New")
		}
	}

	accountStore, err := data.NewAccountStore(db)
	if err != nil {
		return nil, errors.Wrap(err, "NewAccountStore")
	}

	tokenStore, err := data.NewRefreshTokenStore(db, redis, cfg.ErrorReporter, cfg.RefreshTokenTTL)
	if err != nil {
		return nil, errors.Wrap(err, "NewRefreshTokenStore")
	}

	blobStore, err := data.NewBlobStore(cfg.AccessTokenTTL, redis, db, cfg.ErrorReporter)
	if err != nil {
		return nil, errors.Wrap(err, "NewBlobStore")
	}

	keyStore := data.NewRotatingKeyStore()
	if cfg.IdentitySigningKey == nil {
		m := data.NewKeyStoreRotater(
			data.NewEncryptedBlobStore(blobStore, cfg.DBEncryptionKey),
			cfg.AccessTokenTTL,
		)
		err := m.Maintain(keyStore, cfg.ErrorReporter)
		if err != nil {
			return nil, errors.Wrap(err, "Maintain")
		}
	} else {
		keyStore.Rotate(cfg.IdentitySigningKey)
	}

	var actives data.Actives
	if redis != nil {
		actives = dataRedis.NewActives(
			redis,
			cfg.StatisticsTimeZone,
			cfg.DailyActivesRetention,
			cfg.WeeklyActivesRetention,
			5*12,
		)
	}

	oauthProviders := map[string]oauth.Provider{}
	if cfg.GoogleOauthCredentials != nil {
		oauthProviders["google"] = *oauth.NewGoogleProvider(cfg.GoogleOauthCredentials)
	}
	if cfg.GitHubOauthCredentials != nil {
		oauthProviders["github"] = *oauth.NewGitHubProvider(cfg.GitHubOauthCredentials)
	}
	if cfg.FacebookOauthCredentials != nil {
		oauthProviders["facebook"] = *oauth.NewFacebookProvider(cfg.FacebookOauthCredentials)
	}
	if cfg.DiscordOauthCredentials != nil {
		oauthProviders["discord"] = *oauth.NewDiscordProvider(cfg.DiscordOauthCredentials)
	}

	return &App{
		// Provide access to root DB - useful when extending AccountStore functionality
		DB:                db,
		DbCheck:           func() bool { return db.Ping() == nil },
		RedisCheck:        func() bool { return redis != nil && redis.Ping().Err() == nil },
		Config:            cfg,
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
		KeyStore:          keyStore,
		Actives:           actives,
		Reporter:          cfg.ErrorReporter,
		OauthProviders:    oauthProviders,
	}, nil
}
