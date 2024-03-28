package oauth

import (
	"errors"
	"strings"
)

// Credentials is a configuration struct for OAuth Providers
type Credentials struct {
	ID         string
	Secret     string
	Additional map[string]string
}

// NewCredentials parses a credential string in the format
// `id:secret:additional=data...(optional)`
// and returns a Credentials suitable for OAuth Provider configuration.
// If no signing key is provided the default key is used.
// Any content after the third colon is assumed a key-value pair in the form
// `key=value` and is added to the Additional map.
func NewCredentials(credentials string) (*Credentials, error) {
	if strings.Count(credentials, ":") < 1 {
		return nil, errors.New("credentials must be in the format `id:secret:additional=data...(optional)`")
	}
	strs := strings.SplitN(credentials, ":", 3)

	c := &Credentials{
		ID:     strs[0],
		Secret: strs[1],
	}

	if len(strs) == 3 && strs[2] != "" {
		c.Additional = make(map[string]string)
		for _, pair := range strings.Split(strs[2], ":") {
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
