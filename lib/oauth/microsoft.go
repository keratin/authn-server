package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2"
)

// NewMicrosoftProvider returns a AuthN integration for Microsoft OAuth
func NewMicrosoftProvider(credentials *Credentials) *Provider {
	config := &oauth2.Config{
		ClientID:     credentials.ID,
		ClientSecret: credentials.Secret,
		Scopes:       []string{"openid", "user.read", "user.read.all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	return NewProvider(config, func(t *oauth2.Token) (*UserInfo, error) {
		var me struct {
			Id                string `json:"id"`
			UserPrincipalName string `json:"userPrincipalName"`
		}

		client := config.Client(context.TODO(), t)
		resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var user UserInfo
		err = json.Unmarshal(body, &me)
		user.ID = me.Id
		user.Email = me.UserPrincipalName
		fmt.Println(user)
		return &user, err
	}, jose.SigningKey{Key: credentials.SigningKey, Algorithm: jose.HS256})
}
