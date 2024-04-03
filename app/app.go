package app

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/app/data"
	dataRedis "github.com/keratin/authn-server/app/data/redis"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	TOTPCache         data.TOTPCache
	Actives           data.Actives
	Reporter          ops.ErrorReporter
	OauthProviders    map[string]oauth.Provider
	Logger            logrus.FieldLogger
}

func NewApp(cfg *Config, logger logrus.FieldLogger) (*App, error) {
	errorReporter, err := ops.NewErrorReporter(cfg.ErrorReporterCredentials, cfg.ErrorReporterType, logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize error reporter")
	}

	db, err := data.NewDB(cfg.DatabaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "data.NewDB")
	}

	var redis *redis.Client
	if cfg.RedisIsSentinelMode {
		redis, err = dataRedis.NewSentinel(cfg.RedisSentinelMaster, cfg.RedisSentinelNodes, cfg.RedisSentinelPassword)
		if err != nil {
			return nil, errors.Wrap(err, "redis.NewSentinel")
		}
	} else if cfg.RedisURL != nil {
		redis, err = dataRedis.New(cfg.RedisURL)
		if err != nil {
			return nil, errors.Wrap(err, "redis.New")
		}
	}

	accountStore, err := data.NewAccountStore(db)
	if err != nil {
		return nil, errors.Wrap(err, "NewAccountStore")
	}

	tokenStore, err := data.NewRefreshTokenStore(db, redis, errorReporter, cfg.RefreshTokenTTL)
	if err != nil {
		return nil, errors.Wrap(err, "NewRefreshTokenStore")
	}

	blobStore, err := data.NewBlobStore(cfg.AccessTokenTTL, redis, db, errorReporter)
	if err != nil {
		return nil, errors.Wrap(err, "NewBlobStore")
	}

	encryptedBlobStore := data.NewEncryptedBlobStore(blobStore, cfg.DBEncryptionKey)

	keyStore := data.NewRotatingKeyStore()
	if cfg.IdentitySigningKey == nil {
		m := data.NewKeyStoreRotater(
			encryptedBlobStore,
			cfg.AccessTokenTTL,
			logger,
		)
		err := m.Maintain(keyStore, errorReporter)
		if err != nil {
			return nil, errors.Wrap(err, "Maintain")
		}
	} else {
		keyStore.Rotate(cfg.IdentitySigningKey)
	}

	totpCache := data.NewTOTPCache(encryptedBlobStore)

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

	oauthProviders, err := initializeOAuthProviders(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "initializeOAuthProviders")
	}

	return &App{
		// Provide access to root DB - useful when extending AccountStore functionality
		DB:                db,
		DbCheck:           func() bool { return db.Ping() == nil },
		RedisCheck:        func() bool { return redis != nil && redis.Ping(context.TODO()).Err() == nil },
		Config:            cfg,
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
		KeyStore:          keyStore,
		TOTPCache:         totpCache,
		Actives:           actives,
		Reporter:          errorReporter,
		OauthProviders:    oauthProviders,
		Logger:            logger,
	}, nil
}

func initializeOAuthProviders(cfg *Config) (map[string]oauth.Provider, error) {
	oauthProviders := make(map[string]oauth.Provider)
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
	if cfg.MicrosoftOauthCredentials != nil {
		oauthProviders["microsoft"] = *oauth.NewMicrosoftProvider(cfg.MicrosoftOauthCredentials)
	}
	if cfg.AppleOAuthCredentials != nil {
		appleProvider, err := oauth.NewAppleProvider(cfg.AppleOAuthCredentials)
		if err != nil {
			return nil, err
		}
		oauthProviders["apple"] = *appleProvider
	}
	return oauthProviders, nil
}
