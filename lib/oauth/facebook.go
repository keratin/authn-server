package oauth

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"gopkg.in/square/go-jose.v2"
)

// NewFacebookProvider returns a AuthN integration for Facebook OAuth
func NewFacebookProvider(credentials *Credentials, signingKey []byte) *Provider {
	config := &oauth2.Config{
		ClientID:     credentials.ID,
		ClientSecret: credentials.Secret,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}

	return NewProvider(config, func(t *oauth2.Token) (*UserInfo, error) {
		client := config.Client(context.TODO(), t)
		resp, err := client.Get("https://graph.facebook.com/me?fields=id,email")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var user UserInfo
		err = json.Unmarshal(body, &user)
		return &user, err
	}, jose.SigningKey{Key: signingKey, Algorithm: jose.HS256})
}
