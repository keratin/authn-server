package oauth

import (
	"golang.org/x/oauth2"
)

// Provider is a struct wrapping the necessary bits to integrate an OAuth2 provider with AuthN
type Provider struct {
	config   *oauth2.Config
	UserInfo UserInfoFetcher
}

// UserInfo is the minimum necessary needed from an OAuth Provider to connect with AuthN accounts
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// UserInfoFetcher is the function signature for fetching UserInfo from a Provider
type UserInfoFetcher = func(t *oauth2.Token) (*UserInfo, error)

// NewProvider returns a properly configured Provider
func NewProvider(config *oauth2.Config, userInfo UserInfoFetcher) *Provider {
	return &Provider{config, userInfo}
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
