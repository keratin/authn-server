package config

import "time"

type Config struct {
	BcryptCost            int
	UsernameIsEmail       bool
	UsernameMinLength     int
	UsernameDomain        string
	PasswordMinComplexity int
	RefreshTokenTTL       time.Duration
	RedisURL              string
	SessionSigningKey     []byte
}

var oneYear = time.Duration(8766) * time.Hour

func ReadEnv() Config {
	return Config{
		BcryptCost:            11,
		UsernameIsEmail:       true,
		UsernameMinLength:     3,
		UsernameDomain:        "",
		PasswordMinComplexity: 2,
		RefreshTokenTTL:       oneYear,
		RedisURL:              "redis://127.0.0.1:6379/11",
		SessionSigningKey:     []byte("TODO"),
	}
}
