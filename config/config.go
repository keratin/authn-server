package config

type Config struct {
	BcryptCost            int
	UsernameIsEmail       bool
	UsernameMinLength     int
	UsernameDomain        string
	PasswordMinComplexity int
}

func ReadEnv() Config {
	return Config{
		BcryptCost:            11,
		UsernameIsEmail:       true,
		UsernameMinLength:     3,
		UsernameDomain:        "",
		PasswordMinComplexity: 2,
	}
}
