package oauth

import (
	"errors"
	"strings"
)

// Credentials is a configuration struct for OAuth Providers
type Credentials struct {
	ID     string
	Secret string
}

// NewCredentials parses a credential string in the format `id:string` and returns a Credentials
// suitable for OAuth Provider configuration.
func NewCredentials(credentials string) (*Credentials, error) {
	if strings.Count(credentials, ":") != 1 {
		return nil, errors.New("Credentials must be in the format `id:string`")
	}
	strs := strings.SplitN(credentials, ":", 2)
	return &Credentials{
		ID:     strs[0],
		Secret: strs[1],
	}, nil
}
