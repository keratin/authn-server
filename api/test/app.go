package test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
)

func App() *api.App {
	authnURL, err := url.Parse("https://authn.example.com")
	if err != nil {
		panic(err)
	}

	weakKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}

	cfg := config.Config{
		BcryptCost:            4,
		SessionSigningKey:     []byte("TestKey"),
		AuthNURL:              authnURL,
		SessionCookieName:     "authn",
		ApplicationDomains:    []route.Domain{{Hostname: "test.com"}},
		PasswordMinComplexity: 2,
		AppPasswordResetURL:   &url.URL{Scheme: "https", Host: "app.example.com"},
		EnableSignup:          true,
	}

	return &api.App{
		Config:            &cfg,
		KeyStore:          mock.NewKeyStore(weakKey),
		AccountStore:      mock.NewAccountStore(),
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		Actives:           mock.NewActives(),
		Reporter:          &ops.LogReporter{},
	}
}
