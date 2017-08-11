package password_resets

import (
	"fmt"
	"strconv"
	"time"

	"github.com/keratin/authn-server/config"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

const scope = "reset"

type Claims struct {
	Scope string          `json:"scope"`
	Lock  jwt.NumericDate `json:"lock"`
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

func (c *Claims) LockExpired(passwordChangedAt time.Time) bool {
	lockedAt := time.Unix(int64(c.Lock), 0)
	expiredAt := passwordChangedAt.Truncate(time.Second)

	return expiredAt.After(lockedAt)
}

func Parse(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, err
	}

	claims := Claims{}
	err = token.Claims(cfg.ResetSigningKey, &claims)
	if err != nil {
		return nil, err
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
		Time:     time.Now(),
	})
	if err != nil {
		return nil, err
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}

	return &claims, nil
}

func New(cfg *config.Config, accountId int, password_changed_at time.Time) (*Claims, error) {
	return &Claims{
		Scope: scope,
		Lock:  jwt.NewNumericDate(password_changed_at),
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  strconv.Itoa(accountId),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.ResetTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
