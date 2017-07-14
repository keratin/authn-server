package password_resets

import (
	"fmt"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
)

const scope = "reset"

type Claims struct {
	Scope string `json:"scope"`
	Lock  int64  `json:"lock"`
	jwt.StandardClaims
}

func (c *Claims) Sign(hmac_key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(hmac_key)
}

func (c *Claims) LockExpired(password_changed_at *time.Time) bool {
	locked_at := time.Unix(int64(c.Lock), 0)
	expired_at := password_changed_at.Truncate(time.Second)

	return expired_at.After(locked_at)
}

func Parse(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, staticKeyFunc(cfg.ResetSigningKey))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Could not verify JWT")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("JWT is not a Claims")
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}

	err = claims.StandardClaims.Valid()
	if err != nil {
		return nil, err
	}
	if !claims.VerifyAudience(cfg.AuthNURL.String(), true) {
		return nil, fmt.Errorf("token audience not valid")
	}
	if !claims.VerifyIssuer(cfg.AuthNURL.String(), true) {
		return nil, fmt.Errorf("token issuer not valid")
	}

	return claims, nil
}

func New(cfg *config.Config, account_id int, password_changed_at time.Time) (*Claims, error) {
	return &Claims{
		Scope: scope,
		Lock:  password_changed_at.Unix(),
		StandardClaims: jwt.StandardClaims{
			Issuer:    cfg.AuthNURL.String(),
			Subject:   strconv.Itoa(account_id),
			Audience:  cfg.AuthNURL.String(),
			ExpiresAt: time.Now().Add(cfg.ResetTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}, nil
}

func staticKeyFunc(key []byte) func(*jwt.Token) (interface{}, error) {
	return func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	}
}
