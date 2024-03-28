package oauth

import (
	"net/http/httptest"

	"golang.org/x/oauth2"
)

// NewTestProvider returns a special Provider for tests
func NewTestProvider(s *httptest.Server) *Provider {
	return &Provider{
		config: &oauth2.Config{
			ClientID:     "TEST",
			ClientSecret: "SECRET",
			Endpoint: oauth2.Endpoint{
				AuthURL:  s.URL,
				TokenURL: s.URL,
			},
		},
		// The test implementation returns a fake user with an email address copied from the supplied access token.
		UserInfo: func(t *oauth2.Token) (*UserInfo, error) {
			return &UserInfo{
				ID:    t.AccessToken,
				Email: t.AccessToken,
			}, nil
		},
	}
}
