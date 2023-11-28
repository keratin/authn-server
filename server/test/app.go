package test

import (
	"net/http"
	"net/url"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/data/private"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/sirupsen/logrus"
)

func App() *app.App {
	authnURL, err := url.Parse("https://authn.example.com")
	if err != nil {
		panic(err)
	}

	weakKey, err := private.GenerateKey(512)
	if err != nil {
		panic(err)
	}

	cfg := app.Config{
		BcryptCost:              4,
		SessionSigningKey:       []byte("TestKey"),
		AuthNURL:                authnURL,
		SessionCookieName:       "authn",
		OAuthCookieName:         "authn-oauth-nonce",
		DBEncryptionKey:         []byte("DLz2TNDRdWWA5w8YNeCJ7uzcS4WDzQmB"),
		ApplicationDomains:      []route.Domain{{Hostname: "test.com"}},
		PasswordMinComplexity:   2,
		AppPasswordResetURL:     &url.URL{Scheme: "https", Host: "app.example.com"},
		AppPasswordlessTokenURL: &url.URL{Scheme: "https", Host: "app.example.com"},
		EnableSignup:            true,
		SameSite:                http.SameSiteDefaultMode,
		PasswordChangeLogout:    false,
	}

	//Create mock blob stores for the totp cache object (TODO: Create an interface?)
	bs := mock.NewBlobStore(time.Minute, time.Minute)
	ebs := data.NewEncryptedBlobStore(bs, cfg.DBEncryptionKey)

	logger := logrus.New()
	return &app.App{
		Config:            &cfg,
		KeyStore:          mock.NewKeyStore(weakKey),
		AccountStore:      mock.NewAccountStore(),
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		TOTPCache:         data.NewTOTPCache(ebs),
		Actives:           mock.NewActives(),
		Reporter:          &ops.LogReporter{FieldLogger: logger},
		OauthProviders:    map[string]oauth.Provider{},
		Logger:            logger,
	}
}
