package oauth

import (
	"github.com/go-jose/go-jose/v3"
	"golang.org/x/oauth2"
)

// Provider is a struct wrapping the necessary bits to integrate an OAuth2 provider with AuthN
type Provider struct {
	config     *oauth2.Config
	UserInfo   UserInfoFetcher
	signingKey jose.SigningKey
}

// UserInfo is the minimum necessary needed from an OAuth Provider to connect with AuthN accounts
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// UserInfoFetcher is the function signature for fetching UserInfo from a Provider
type UserInfoFetcher = func(t *oauth2.Token) (*UserInfo, error)

// NewProvider returns a properly configured Provider
func NewProvider(config *oauth2.Config, userInfo UserInfoFetcher, signingKey jose.SigningKey) *Provider {
	return &Provider{config: config, UserInfo: userInfo, signingKey: signingKey}
}

// Config returns a complete oauth2.Config after injecting the RedirectURL
func (p *Provider) Config(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Scopes:       p.config.Scopes,
		Endpoint:     p.config.Endpoint,
		RedirectURL:  redirectURL,
	}
}

func (p *Provider) SigningKey() jose.SigningKey {
	//TODO: allow override with dynamic calc for apple
	return p.signingKey
}
