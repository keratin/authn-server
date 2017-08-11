package sessions

import (
	"time"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

type Claims struct {
	Azp string `json:"azp"`
	jwt.Claims
}

func (c *Claims) Sign(hmacKey []byte) (string, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmacKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

func Parse(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, err
	}

	claims := Claims{}
	err = token.Claims(cfg.SessionSigningKey, &claims)
	if err != nil {
		return nil, err
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
	})
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

func New(store data.RefreshTokenStore, cfg *config.Config, accountID int) (*Claims, error) {
	refreshToken, err := store.Create(accountID)
	if err != nil {
		return nil, err
	}

	return &Claims{
		Azp: "", // TODO: audience
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  string(refreshToken),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
