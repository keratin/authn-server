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
	Additional map[string]string
}

// NewCredentials parses a credential string in the format
// `id:string:signing_key(optional):additional=data...(optional)`
// and returns a Credentials suitable for OAuth Provider configuration.
// If no signing key is provided the default key is used.
// Any content after the third colon is assumed a key-value pair in the form
// `key=value` and is added to the Additional map.
func NewCredentials(credentials string, defaultKey []byte) (*Credentials, error) {
	if strings.Count(credentials, ":") < 1 {
		return nil, errors.New("credentials must be in the format `id:string:signing_key(optional):additional=data...(optional)`")
	}
	strs := strings.SplitN(credentials, ":", 4)

	c := &Credentials{
		ID:     strs[0],
		Secret: strs[1],
	}

	if len(strs) >= 3 && strs[2] != "" {
		key, err := hex.DecodeString(strs[2])
		if err != nil {
			return nil, fmt.Errorf("failed to decode signing key: %w", err)
		}
		c.SigningKey = key
	} else {
		c.SigningKey = defaultKey
	}

	if len(strs) == 4 && strs[3] != "" {
		c.Additional = make(map[string]string)
		for _, pair := range strings.Split(strs[3], ":") {
			if pair == "" {
				continue
			}
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 1 {
				c.Additional[kv[0]] = ""
			} else {
				c.Additional[kv[0]] = kv[1]
			}
		}
	}
	return c, nil
}
