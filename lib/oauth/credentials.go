package oauth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Credentials is a configuration struct for OAuth Providers
type Credentials struct {
	ID         string
	Secret     string
	SigningKey []byte
}

// NewCredentials parses a credential string in the format `id:string:signing_key(optional)`
// and returns a Credentials suitable for OAuth Provider configuration.  If no signing key is
// provided the default key is used.
func NewCredentials(credentials string, defaultKey []byte) (*Credentials, error) {
	if strings.Count(credentials, ":") < 1 {
		return nil, errors.New("Credentials must be in the format `id:string:signing_key(optional)`")
	}
	strs := strings.SplitN(credentials, ":", 3)

	c := &Credentials{
		ID:     strs[0],
		Secret: strs[1],
	}

	if len(strs) == 3 {
		key, err := hex.DecodeString(strs[2])
		if err != nil {
			return nil, fmt.Errorf("failed to decode signing key: %w", err)
		}
		c.SigningKey = key
	} else {
		c.SigningKey = defaultKey
	}
	return c, nil
}
