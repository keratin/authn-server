package passwordless

import (
	"fmt"
	"strconv"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

const scope = "passwordless"

type Claims struct {
	Scope string `json:"scope"`
	jwt.Claims
}

func (c *Claims) Sign(hmacKey []byte) (string, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmacKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", errors.Wrap(err, "NewSigner")
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

func Parse(tokenStr string, cfg *app.Config) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "ParseSigned")
	}

	claims := Claims{}
	err = token.Claims(cfg.PasswordlessTokenSigningKey, &claims)
	if err != nil {
		return nil, errors.Wrap(err, "Claims")
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
		Time:     time.Now(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Validate")
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}

	return &claims, nil
}

func New(cfg *app.Config, accountID int) (*Claims, error) {
	return &Claims{
		Scope: scope,
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  strconv.Itoa(accountID),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.PasswordlessTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
