package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"net/url"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

func derive(base []byte, salt string) []byte {
	return pbkdf2.Key(base, []byte(salt), 2e5, 64, sha256.New)
}

type Config struct {
	BcryptCost            int
	UsernameIsEmail       bool
	UsernameMinLength     int
	UsernameDomain        string
	PasswordMinComplexity int
	RefreshTokenTTL       time.Duration
	RedisURL              *url.URL
	DatabaseURL           *url.URL
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

	secretBase := make([]byte, 64)
	_, err = rand.Read(secretBase)
	if err != nil {
		panic(err)
	}

	dbUrl, err := url.Parse("sqlite3://localhost/dev")
	if err != nil {
		panic(err)
	}

	redisURL, err := url.Parse("redis://127.0.0.1:6379/11")
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
		RedisURL:              redisURL,
		SessionSigningKey:     derive(secretBase, "session-key-salt"),
		IdentitySigningKey:    identityKey,
		AuthNURL:              authnUrl,
		MountedPath:           authnUrl.Path,
		ForceSSL:              authnUrl.Scheme == "https",
		AccessTokenTTL:        time.Hour,
		DatabaseURL:           dbUrl,
	}
}
