package app

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ErrMissingEnvVar string

var supportedSchemes = map[string]struct{}{
	"http":     {},
	"https":    {},
	"mysql":    {},
	"sqlite3":  {},
	"postgres": {},
	"redis":    {},
	"rediss":   {},
}

func (name ErrMissingEnvVar) Error() string {
	return "missing environment variable: " + string(name)
}

func requireEnv(name string) (string, error) {
	if val, ok := os.LookupEnv(name); ok {
		return val, nil
	}

	return "", ErrMissingEnvVar(name)
}

func lookupInt(name string, def int) (int, error) {
	if val, ok := os.LookupEnv(name); ok {
		return strconv.Atoi(val)
	}

	return def, nil
}

func lookupBool(name string, def bool) (bool, error) {
	if val, ok := os.LookupEnv(name); ok {
		return regexp.MatchString("^(?i:t|true|yes)$", val)
	}

	return def, nil
}

func LookupURL(name string) (*url.URL, error) {
	if val, ok := os.LookupEnv(name); ok {
		url, err := url.ParseRequestURI(val)
		if err == nil {
			if _, ok := supportedSchemes[url.Scheme]; !ok {
				return nil, fmt.Errorf("unsupported URL: %v", val)
			}
			return url, nil
		}
		return nil, err
	}
	return nil, nil
}

// lookupOAuthSigningKey checks environment for a hex-encoded key configured for the given provider, eg "GOOGLE_OAUTH_SIGNING_KEY".
func lookupOAuthSigningKey(providerName string) ([]byte, error) {
	keyString, ok := os.LookupEnv(fmt.Sprintf("%s_OAUTH_SIGNING_KEY", strings.ToUpper(providerName)))
	if ok {
		return hex.DecodeString(keyString)
	}
	return nil, fmt.Errorf("missing environment variable: %s_OAUTH_SIGNING_KEY", strings.ToUpper(providerName))
}
