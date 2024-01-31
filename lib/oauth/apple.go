package oauth

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/keratin/authn-server/lib/oauth/apple"
	"golang.org/x/oauth2"
)

// NewAppleProvider returns a AuthN integration for sign-in with Apple OAuth
func NewAppleProvider(credentials *Credentials) (*Provider, error) {
	config := &oauth2.Config{
		ClientID: credentials.ID,
		// ClientSecret for apple is built using apple.GenerateSecret
		// this function is passed to the provider for use as an override
		// and built fresh on each call to provider.Config(returnURL).
		ClientSecret: "",
		Scopes:       []string{"email"},
		Endpoint:     apple.Endpoint(),
	}

	teamID, keyID, expiresIn, constructErr := apple.ExtractCredentialData(credentials.Additional)
	if constructErr != nil {
		return nil, fmt.Errorf("apple: failed to extract credentials: %w", constructErr)
	}

	keyBytes, err := hex.DecodeString(credentials.Secret)
	if err != nil {
		return nil, fmt.Errorf("apple: failed to decode key from client secret: %w", err)
	}

	signingKey, constructErr := apple.ParsePrivateKey(keyBytes, keyID)
	if constructErr != nil {
		return nil, fmt.Errorf("apple: failed to parse signing key: %w", constructErr)
	}

	appleTokenReader := apple.NewTokenReader(config.ClientID)

	getAppleUserInfo := func(t *oauth2.Token) (*UserInfo, error) {
		id, email, err := appleTokenReader.GetUserDetailsFromToken(t)

		if err != nil {
			return nil, err
		}

		return &UserInfo{
			ID:    id,
			Email: email,
		}, nil
	}

	return NewProvider(config, getAppleUserInfo,
		WithSecretGenerator(func() (string, error) {
			return apple.GenerateSecret(*signingKey, keyID, config.ClientID, teamID, expiresIn)
		}),
		// Apple requires form_post response mode if scopes are requested
		WithAuthCodeOptions(oauth2.SetAuthURLParam("response_mode", "form_post")),
		// So we need to handle returns via POST instead of GET
		WithReturnMethod(http.MethodPost)), nil
}
