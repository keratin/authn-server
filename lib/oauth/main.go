package oauth

import "golang.org/x/oauth2"

// Provider is the minimum interface necessary to integrate an OAuth2 provider with AuthN
type Provider interface {
	// Config returns the proper oauth2.Config for a Provider
	Config() *oauth2.Config
	// UserInfo uses the oauth2.Token to fetch id and email from an appropriate endpoint
	UserInfo(tok *oauth2.Token) (*UserInfo, error)
}

// UserInfo is the minimum necessary needed from an OAuth Provider to connect with AuthN accounts
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
