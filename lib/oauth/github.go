package oauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// NewGitHubProvider returns a AuthN integration for GitHub OAuth
func NewGitHubProvider(credentials *Credentials) *Provider {
	config := &oauth2.Config{
		ClientID:     credentials.ID,
		ClientSecret: credentials.Secret,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	getPrimaryEmail := func(t *oauth2.Token) (string, error) {
		client := config.Client(context.TODO(), t)
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var emails []struct {
			Email   string
			Primary bool
		}
		err = json.Unmarshal(body, &emails)
		if err != nil {
			return "", err
		}
		for _, email := range emails {
			if email.Primary {
				return email.Email, nil
			}
		}
		return "", nil
	}

	getID := func(t *oauth2.Token) (string, error) {
		client := config.Client(context.TODO(), t)
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var user struct {
			ID int
		}
		err = json.Unmarshal(body, &user)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(user.ID), nil
	}

	return &Provider{
		config: config,
		UserInfo: func(t *oauth2.Token) (*UserInfo, error) {
			id, err := getID(t)
			if err != nil {
				return nil, err
			}

			email, err := getPrimaryEmail(t)
			if err != nil {
				return nil, err
			}

			return &UserInfo{
				ID:    id,
				Email: email,
			}, nil
		},
	}
}
