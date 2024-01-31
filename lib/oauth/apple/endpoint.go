package apple

import (
	"fmt"

	"golang.org/x/oauth2"
)

const BaseURL = "https://appleid.apple.com"

func Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:   fmt.Sprintf("%s/auth/authorize", BaseURL),
		TokenURL:  fmt.Sprintf("%s/auth/token", BaseURL),
		AuthStyle: 0,
	}
}
