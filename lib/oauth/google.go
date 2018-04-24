package oauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
	_google "golang.org/x/oauth2/google"
)

// Google implements oauth.Provider for AuthN
type Google struct {
}

// Config reads the configuration for the Google provider from ENV variables
func (g *Google) Config() *oauth2.Config {
	id, _ := os.LookupEnv("GOOGLE_OAUTH_ID")
	secret, _ := os.LookupEnv("GOOGLE_OAUTH_SECRET")
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       []string{"email"},
		// TODO: inject or append
		RedirectURL: "http://localhost:7001/oauth/google/return",
		Endpoint:    _google.Endpoint,
	}
}

// UserInfo queries Google to fetch user information from an OAuth token.
func (g *Google) UserInfo(t *oauth2.Token) (*UserInfo, error) {
	client := g.Config().Client(context.TODO(), t)
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

	var user UserInfo
	err = json.Unmarshal(body, &user)
	return &user, err
}
