package app

type configurer func(c *Config) error

func configure(fns []configurer) (*Config, error) {
	var err error
	c := Config{
		UsernameMinLength: 3,
		SessionCookieName: "authn",
		OAuthCookieName:   "authn-oauth-nonce",
	}
	for _, fn := range fns {
		err = fn(&c)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}
