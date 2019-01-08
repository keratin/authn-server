package sessions

import (
	"fmt"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

const scope = "refresh"

type Claims struct {
	Scope string `json:"scope"`
	Azp   string `json:"azp"`
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
	err = token.Claims(cfg.SessionSigningKey, &claims)
	if err != nil {
		return nil, errors.Wrap(err, "Claims")
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Validate")
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}

	return &claims, nil
}

func New(store data.RefreshTokenStore, cfg *app.Config, accountID int, authorizedAudience string) (*Claims, error) {
	refreshToken, err := store.Create(accountID)
	if err != nil {
		return nil, errors.Wrap(err, "Create")
	}

	return &Claims{
		Scope: scope,
		Azp:   authorizedAudience,
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  string(refreshToken),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
