package tokens_test

import (
	"net/url"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data/mock"
	"github.com/keratin/authn/tests"
	"github.com/keratin/authn/tokens"
)

func TestSessionJWT(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	cfg := config.Config{
		AuthNURL:          &url.URL{Scheme: "http", Host: "authn.example.com"},
		SessionSigningKey: []byte("key-a-reno"),
	}

	session, err := tokens.NewSessionJWT(store, &cfg, 658908)
	if err != nil {
		t.Fatal(err)
	}
	tests.AssertEqual(t, "http://authn.example.com", session.Issuer)
	tests.AssertEqual(t, "http://authn.example.com", session.Audience)
	tests.AssertEqual(t, "RefreshToken:658908", session.Subject)
	tests.AssertEqual(t, "", session.Azp)
	tests.RefuteEqual(t, int64(0), session.IssuedAt)

	sessionString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, session).SignedString(cfg.SessionSigningKey)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := tokens.ParseSessionJWT(sessionString, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	tests.AssertEqual(t, "http://authn.example.com", claims.Issuer)
	tests.AssertEqual(t, "http://authn.example.com", claims.Audience)
	tests.AssertEqual(t, "RefreshToken:658908", claims.Subject)
	tests.AssertEqual(t, "", claims.Azp)
	tests.RefuteEqual(t, int64(0), claims.IssuedAt)
}

func TestParseInvalidSessionJWT(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	authn := url.URL{Scheme: "http", Host: "authn.example.com"}
	app := url.URL{Scheme: "http", Host: "app.example.com"}
	key := []byte("current key")
	oldKey := []byte("old key")

	invalids := []string{}
	var token *tokens.SessionClaims
	var tokenStr string
	var cfg config.Config
	var err error

	// This invalid JWT was signed with an old key.
	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: oldKey}
	token, err = tokens.NewSessionJWT(store, &cfg, 1)
	errIsFatal(t, err)
	tokenStr, err = jwt.NewWithClaims(jwt.SigningMethodHS256, token).SignedString(cfg.SessionSigningKey)
	errIsFatal(t, err)
	invalids = append(invalids, tokenStr)

	// This invalid JWT was signed for a different audience.
	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: key}
	token, err = tokens.NewSessionJWT(store, &cfg, 2)
	errIsFatal(t, err)
	token.Audience = app.String()
	tokenStr, err = jwt.NewWithClaims(jwt.SigningMethodHS256, token).SignedString(cfg.SessionSigningKey)
	errIsFatal(t, err)
	invalids = append(invalids, tokenStr)

	// This invalid JWT was signed with "none" alg
	cfg = config.Config{AuthNURL: &authn}
	token, err = tokens.NewSessionJWT(store, &cfg, 3)
	errIsFatal(t, err)
	tokenStr, err = jwt.NewWithClaims(jwt.SigningMethodNone, token).SignedString(jwt.UnsafeAllowNoneSignatureType)
	errIsFatal(t, err)
	invalids = append(invalids, tokenStr)

	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: key}
	for i, invalid := range invalids {
		_, err := tokens.ParseSessionJWT(invalid, &cfg)
		if err == nil {
			t.Errorf("invalid token [%v] was parsed as valid", i)
		}
	}
}

func errIsFatal(t *testing.T, err error) {
	if err != nil {
		panic(err)
	}
}
