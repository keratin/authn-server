package oauth

import (
	"net/http"

	"golang.org/x/oauth2"
)

// Provider is a struct wrapping the necessary bits to integrate an OAuth2 provider with AuthN
type Provider struct {
	config          *oauth2.Config
	UserInfo        UserInfoFetcher
	secretGenerator SecretGenerator
	authCodeOptions []oauth2.AuthCodeOption
	returnMethod    string
}

// UserInfo returns the information needed from an OAuth Provider to connect with AuthN accounts
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// UserInfoFetcher is the function signature for fetching UserInfo from a Provider
type UserInfoFetcher = func(t *oauth2.Token) (*UserInfo, error)

type override func(*Provider)

// SecretGenerator is the function signature for calculating a dynamic client secret.
type SecretGenerator = func() (string, error)

func WithSecretGenerator(so SecretGenerator) override {
	return func(p *Provider) {
		p.secretGenerator = so
	}
}

// WithAuthCodeOptions sets additional options to use when requesting an auth code.
func WithAuthCodeOptions(ao ...oauth2.AuthCodeOption) override {
	return func(p *Provider) {
		p.authCodeOptions = ao
	}
}

// WithReturnMethod sets the HTTP method to use when returning from the OAuth provider.
func WithReturnMethod(method string) override {
	return func(p *Provider) {
		p.returnMethod = method
	}
}

// NewProvider returns a properly configured Provider
func NewProvider(config *oauth2.Config, userInfo UserInfoFetcher, overrides ...override) *Provider {
	p := &Provider{config: config, UserInfo: userInfo}

	for _, o := range overrides {
		o(p)
	}

	return p
}

// Config returns a complete oauth2.Config after injecting the RedirectURL
func (p *Provider) Config(redirectURL string) (*oauth2.Config, error) {
	secret, err := p.Secret()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: secret,
		Scopes:       p.config.Scopes,
		Endpoint:     p.config.Endpoint,
		RedirectURL:  redirectURL,
	}, nil
}

func (p *Provider) Secret() (string, error) {
	if p.secretGenerator != nil {
		return p.secretGenerator()
	}
	return p.config.ClientSecret, nil
}

func (p *Provider) AuthCodeOptions() []oauth2.AuthCodeOption {
	if p.authCodeOptions != nil {
		return p.authCodeOptions
	}
	return nil
}

func (p *Provider) ReturnMethod() string {
	if p.returnMethod != "" {
		return p.returnMethod
	}
	return http.MethodGet
}
