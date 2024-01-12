package app

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
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
