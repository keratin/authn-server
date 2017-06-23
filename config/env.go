package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func requireEnv(name string) (string, error) {
	if val, ok := os.LookupEnv(name); ok {
		return val, nil
	} else {
		return "", fmt.Errorf("Missing environment variable: %s. See https://github.com/keratin/authn/wiki/Server-Configuration for details.", name)
	}
}

func lookupInt(name string, def int) (int, error) {
	if val, ok := os.LookupEnv(name); ok {
		return strconv.Atoi(val)
	} else {
		return def, nil
	}
}

func lookupBool(name string, def bool) (bool, error) {
	if val, ok := os.LookupEnv(name); ok {
		return regexp.MatchString("^(?i:t|true|yes)$", val)
	} else {
		return def, nil
	}
}
