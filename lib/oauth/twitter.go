package oauth

import (
	"context"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
)

// NewTwitterProvider returns a AuthN integration for Twitter OAuth
func NewTwitterProvider(credentials *Credentials) *Provider {
	config := &oauth2.Config{
		ClientID:     credentials.ID,
		ClientSecret: credentials.Secret,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
	}

	return NewProvider(config, func(t *oauth2.Token) (*UserInfo, error) {
		client := config.Client(context.TODO(), t)
		resp, err := client.Get("https://api.twitter.com/2/me")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var user UserInfo
		err = json.Unmarshal(body, &user)
		return &user, err
	})
}
