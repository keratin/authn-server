package test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
)

func App() *api.App {
	accountStore := mock.NewAccountStore()

	authnUrl, err := url.Parse("https://authn.example.com")
	if err != nil {
		panic(err)
	}

	weakKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}

	keyStore := mock.NewKeyStore(weakKey)

	cfg := config.Config{
		BcryptCost:            4,
		SessionSigningKey:     []byte("TODO"),
		AuthNURL:              authnUrl,
		SessionCookieName:     "authn",
		ApplicationDomains:    []string{"test.com"},
		PasswordMinComplexity: 2,
		AppPasswordResetURL:   &url.URL{Scheme: "https", Host: "app.example.com"},
		EnableSignup:          true,
	}

	tokenStore := mock.NewRefreshTokenStore()

	actives := mock.NewActives()

	return &api.App{
		KeyStore:          keyStore,
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
		Config:            &cfg,
		Actives:           actives,
	}
}
