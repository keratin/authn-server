package config

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"time"
)

type Config struct {
	BcryptCost            int
	UsernameIsEmail       bool
	UsernameMinLength     int
	UsernameDomain        string
	PasswordMinComplexity int
	RefreshTokenTTL       time.Duration
	RedisURL              string
	SessionSigningKey     []byte
	IdentitySigningKey    *rsa.PrivateKey
	AuthNURL              *url.URL
	ForceSSL              bool
	MountedPath           string
	AccessTokenTTL        time.Duration
}

var oneYear = time.Duration(8766) * time.Hour

func ReadEnv() Config {
	authnUrl, err := url.Parse("https://example.com/authn")
	if err != nil {
		panic(err)
	}

	identityKey, err := rsa.GenerateKey(rand.Reader, 2056)
	if err != nil {
		panic(err)
	}

	return Config{
		BcryptCost:            11,
		UsernameIsEmail:       true,
		UsernameMinLength:     3,
		UsernameDomain:        "",
		PasswordMinComplexity: 2,
		RefreshTokenTTL:       oneYear,
		RedisURL:              "redis://127.0.0.1:6379/11",
		SessionSigningKey:     []byte("TODO"),
		IdentitySigningKey:    identityKey,
		AuthNURL:              authnUrl,
		MountedPath:           authnUrl.Path,
		ForceSSL:              authnUrl.Scheme == "https",
		AccessTokenTTL:        time.Hour,
	}
}
