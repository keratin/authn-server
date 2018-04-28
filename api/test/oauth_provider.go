package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/keratin/authn-server/lib/oauth"
	"golang.org/x/oauth2"
)

type OauthProvider struct {
	server *httptest.Server
}

func NewOauthProvider(s *httptest.Server) *OauthProvider {
	return &OauthProvider{server: s}
}

func ProviderApp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		j, _ := json.Marshal(map[string]interface{}{
			"access_token":  r.FormValue("code"),
			"refresh_token": "REFRESHTOKEN",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(j)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (p *OauthProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "TEST",
		ClientSecret: "SECRET",
		Scopes:       []string{"email"},
		// TODO: redirect URL should be injected from app (routing table)
		RedirectURL: "http://localhost:7001/oauth/test/return",
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.server.URL,
			TokenURL: p.server.URL,
		},
	}
}

// UserInfo returns a fake user with an email address copied from the supplied access token.
func (p *OauthProvider) UserInfo(t *oauth2.Token) (*oauth.UserInfo, error) {
	return &oauth.UserInfo{
		ID:    t.AccessToken,
		Email: t.AccessToken,
	}, nil
}
