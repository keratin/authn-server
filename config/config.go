package config

type Config struct {
	BcryptCost int
}

func ReadEnv() Config {
	return Config{
		BcryptCost: 11,
	}
}
