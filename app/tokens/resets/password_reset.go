package resets

import (
	"fmt"
	"strconv"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

const scope = "reset"

type Claims struct {
	Scope string          `json:"scope"`
	Lock  *jwt.NumericDate `json:"lock"`
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

func (c *Claims) LockExpired(passwordChangedAt time.Time) bool {
	lockedAt := time.Unix(int64(*c.Lock), 0)
	expiredAt := passwordChangedAt.Truncate(time.Second)

	return expiredAt.After(lockedAt)
}

func Parse(tokenStr string, cfg *app.Config) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "ParseSigned")
	}

	claims := Claims{}
	err = token.Claims(cfg.ResetSigningKey, &claims)
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

func New(cfg *app.Config, accountID int, passwordChangedAt time.Time) (*Claims, error) {
	return &Claims{
		Scope: scope,
		Lock:  jwt.NewNumericDate(passwordChangedAt),
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  strconv.Itoa(accountID),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.ResetTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
