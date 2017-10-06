package config

// declare name & type/parser (flag.Value?)
// configure flags (parse ENV into the default)
// parse flags (and return remaining args?) into inputs
// validate inputs
// set config.Config from inputs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/keratin/authn-server/lib/route"
	"github.com/namsral/flag"
)

type address url.URL

func (a *address) String() string {
	return a.Get().String()
}

func (a *address) Set(val string) error {
	parsed, err := url.Parse(val)
	if err == nil {
		*a = address(*parsed)
	}
	return err
}

func (a address) Get() *url.URL {
	u := url.URL(a)
	return &u
}

type seconds int

func (s *seconds) String() string {
	return fmt.Sprintf("%ds", s)
}

func (s *seconds) Set(val string) error {
	i, err := strconv.Atoi(val)
	if err == nil {
		*s = seconds(i)
	}
	return err
}

func (s seconds) Get() time.Duration {
	return time.Duration(s) * time.Second
}

type strs []string

func (ss *strs) String() string {
	return strings.Join(*ss, ", ")
}

func (ss *strs) Set(val string) error {
	*ss = strs(strings.Split(val, ","))
	return nil
}

func (ss strs) Get() []string {
	return []string(ss)
}

type domains []route.Domain

func (ds *domains) String() string {
	var strs []string
	for _, d := range *ds {
		strs = append(strs, d.String())
	}
	return strings.Join(strs, ",")
}

func (ds *domains) Set(val string) error {
	for _, d := range strings.Split(val, ",") {
		*ds = append(*ds, route.ParseDomain(d))
	}
	return nil
}

func (ds domains) Get() []route.Domain {
	return []route.Domain(ds)
}

type rsaKey rsa.PrivateKey

func (k *rsaKey) String() string {
	pk := rsa.PrivateKey(*k)
	if len(pk.Primes) == 0 {
		return ""
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(&pk),
	}))
}

func (k *rsaKey) Set(val string) error {
	val = strings.Replace(val, `\n`, "\n", -1)
	block, _ := pem.Decode([]byte(val))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		*k = rsaKey(*key)
	}
	return err
}

func (k rsaKey) Get() *rsa.PrivateKey {
	pk := rsa.PrivateKey(k)
	return &pk
}

type timeLocation time.Location

func (tz *timeLocation) String() string {
	return tz.Get().String()
}

func (tz *timeLocation) Set(val string) error {
	loc, err := time.LoadLocation(val)
	if err == nil {
		*tz = timeLocation(*loc)
	}
	return err
}

func (tz timeLocation) Get() *time.Location {
	l := time.Location(tz)
	return &l
}

var authnURL address
var redisURL address
var databaseURL address
var appPasswordChangeURL address
var appPasswordResetURL address
var appDomains domains
var secretKeyBase string
var bcryptCost int
var passwordPolicyScore int
var usernameIsEmail bool
var enableSignup bool
var emailUsernameDomains strs
var refreshTokenTTL seconds
var passwordResetTokenTTL seconds
var accessTokenTTL seconds
var httpAuthUsername string
var httpAuthPassword string
var rsaPrivateKey rsaKey
var timeZone timeLocation
var dailyActivesRetention int
var weeklyActivesRetention int
var sentryDSN string

func init() {
	// TODO: descriptions
	flag.Var(&authnURL, "authn_url", "The issuer of ID tokens. Must be a URL that the application backend can resolve in order to fetch AuthN's public key for JWT verification.\n\nIf the URL includes a path, all API routes will be relative to it.")
	flag.Var(&redisURL, "redis_url", "")
	flag.Var(&databaseURL, "database_url", "")
	flag.Var(&appPasswordChangeURL, "app_password_change_url", "")
	flag.Var(&appPasswordResetURL, "app_password_reset_url", "")
	flag.Var(&appDomains, "app_domains", "The list of domains that may refer traffic and be valid JWT audiences. If the domain includes a port, it must match referred traffic. If the domain does not include a port, it will match any referred traffic port. Ports 80 and 443 are matched against schemes.")
	flag.StringVar(&secretKeyBase, "secret_key_base", "", "")
	flag.IntVar(&bcryptCost, "bcrypt_cost", 11, "10-15")
	flag.IntVar(&passwordPolicyScore, "password_policy_score", 2, "0-5")
	flag.BoolVar(&usernameIsEmail, "username_is_email", false, "")
	flag.BoolVar(&enableSignup, "enable_signup", true, "")
	flag.Var(&emailUsernameDomains, "email_username_domains", "")
	refreshTokenTTL.Set(strconv.Itoa(86400 * 365.25))
	flag.Var(&refreshTokenTTL, "refresh_token_ttl", "")
	passwordResetTokenTTL.Set(strconv.Itoa(1800))
	flag.Var(&passwordResetTokenTTL, "password_reset_token_ttl", "")
	accessTokenTTL.Set(strconv.Itoa(3600))
	flag.Var(&accessTokenTTL, "access_token_ttl", "")
	flag.StringVar(&httpAuthUsername, "http_auth_username", randStr(), "")
	flag.StringVar(&httpAuthPassword, "http_auth_password", randStr(), "")
	flag.Var(&rsaPrivateKey, "rsa_private_key", "")
	timeZone.Set("UTC")
	flag.Var(&timeZone, "time_zone", "")
	flag.IntVar(&dailyActivesRetention, "daily_actives_retention", 365, "")
	flag.IntVar(&weeklyActivesRetention, "weekly_actives_retention", 104, "")
	flag.StringVar(&sentryDSN, "sentry_dsn", "", "")
}

func ReadFlags() (*Config, error) {
	if !flag.Parsed() {
		return nil, fmt.Errorf("flags not parsed")
	}

	if len(appDomains) == 0 {
		return nil, fmt.Errorf("missing variable: %s", "APP_DOMAINS")
	}
	if len(secretKeyBase) == 0 {
		return nil, fmt.Errorf("missing variable: %s", "SECRET_KEY_BASE")
	}
	if bcryptCost < 10 {
		return nil, fmt.Errorf("missing variable: %s", "BCRYPT_COST")
	}
	if passwordPolicyScore < 0 || passwordPolicyScore > 4 {
		return nil, fmt.Errorf("missing variable: %s", "PASSWORD_POLICY_SCORE")
	}
	if (authnURL == address{}) {
		return nil, fmt.Errorf("missing variable: %s", "AUTHN_URL")
	}
	if (databaseURL == address{}) {
		return nil, fmt.Errorf("missing variable: %s", "DATABASE_URL")
	}
	if (redisURL == address{}) {
		return nil, fmt.Errorf("missing variable: %s", "REDIS_URL")
	}

	return &Config{
		AccessTokenTTL:         accessTokenTTL.Get(),
		ApplicationDomains:     appDomains,
		AppPasswordChangedURL:  appPasswordChangeURL.Get(),
		AppPasswordResetURL:    appPasswordResetURL.Get(),
		AuthNURL:               authnURL.Get(),
		AuthPassword:           httpAuthPassword,
		AuthUsername:           httpAuthUsername,
		BcryptCost:             bcryptCost,
		DailyActivesRetention:  dailyActivesRetention,
		DatabaseURL:            databaseURL.Get(),
		DBEncryptionKey:        derive([]byte(secretKeyBase), "db-encryption-key-salt")[:32],
		EnableSignup:           enableSignup,
		ForceSSL:               authnURL.Scheme == "https",
		IdentitySigningKey:     rsaPrivateKey.Get(),
		MountedPath:            authnURL.Path,
		PasswordMinComplexity:  passwordPolicyScore,
		RedisURL:               redisURL.Get(),
		RefreshTokenTTL:        refreshTokenTTL.Get(),
		ResetSigningKey:        derive([]byte(secretKeyBase), "password-reset-token-key-salt"),
		ResetTokenTTL:          passwordResetTokenTTL.Get(),
		SentryDSN:              sentryDSN,
		SessionCookieName:      "authn",
		SessionSigningKey:      derive([]byte(secretKeyBase), "session-key-salt"),
		StatisticsTimeZone:     timeZone.Get(),
		UsernameDomains:        emailUsernameDomains,
		UsernameIsEmail:        usernameIsEmail,
		UsernameMinLength:      3,
		WeeklyActivesRetention: weeklyActivesRetention,
	}, nil
}

func randStr() string {
	i, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err != nil {
		panic(err)
	}
	return i.String()
}
