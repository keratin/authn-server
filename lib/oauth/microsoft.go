package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
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

	return &Provider{
		config: config,
		UserInfo: func(t *oauth2.Token) (*UserInfo, error) {
			var me struct {
				id                string
				userPrincipalName string
			}

			client := config.Client(context.TODO(), t)
			resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			fmt.Println(string(body))

			var user UserInfo
			err = json.Unmarshal(body, &me)
			user.ID = me.id
			user.Email = me.userPrincipalName
			fmt.Println(user)
			return &user, err
		},
	}
}
