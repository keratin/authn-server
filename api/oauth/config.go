package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
	_google "golang.org/x/oauth2/google"
)

type provider interface {
	config() *oauth2.Config
	userInfo(tok *oauth2.Token) (*userInfo, error)
}

type userInfo struct {
	id    string
	email string
	name  string
}

type google struct {
}

func (g *google) config() *oauth2.Config {
	id, _ := os.LookupEnv("GOOGLE_OAUTH_ID")
	secret, _ := os.LookupEnv("GOOGLE_OAUTH_SECRET")
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       []string{"email"},
		RedirectURL:  "http://localhost:7001/oauth/google/return",
		Endpoint:     _google.Endpoint,
	}
}

func (g *google) userInfo(t *oauth2.Token) (*userInfo, error) {
	client := g.config().Client(context.TODO(), t)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// {
	//   "id": "1234567890",
	//   "email": "user@example.com",
	//   "verified_email": true,
	//   "name": "Example User",
	//   "given_name": "Example",
	//   "family_name": "User",
	//   "link": "https://plus.google.com/1234567890",
	//   "picture": "https://lh6.googleusercontent.com/path/to/photo.jpg"
	// }
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user userInfo
	err = json.Unmarshal(body, user)
	return &user, err
}

func getProvider(provider string) (provider, error) {
	if provider == "google" {
		return &google{}, nil
	}
	return nil, fmt.Errorf("unknown provider " + provider)
}
