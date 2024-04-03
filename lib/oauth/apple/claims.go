package apple

import (
	"fmt"

	"github.com/go-jose/go-jose/v3/jwt"
)

type Claims struct {
	Email string `json:"email"`
	jwt.Claims
}

// Validate performs apple-specific id_token validation.
// `email` is the only additional claim we currently require.
// See https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/authenticating_users_with_sign_in_with_apple#3383773
// for more details.
func (c Claims) Validate(clientID string) error {
	if clientID == "" {
		return fmt.Errorf("cannot validate with empty clientID")
	}

	if c.Email == "" {
		return fmt.Errorf("missing claim 'email'")
	}

	if c.Expiry == nil {
		return fmt.Errorf("missing claim 'exp'")
	}

	if c.IssuedAt == nil {
		return fmt.Errorf("missing claim 'iat'")
	}

	return c.Claims.Validate(jwt.Expected{
		Issuer:   BaseURL,
		Audience: jwt.Audience{clientID},
	})
}
