package api

import (
	"net/url"

	"github.com/keratin/authn-server/lib/route"
)

func OriginValidator(domains []route.Domain) func(string) bool {
	return func(origin string) bool {
		originURL, err := url.Parse(origin)
		if err != nil {
			return false
		}

		for _, d := range domains {
			if d.Matches(originURL) {
				return true
			}
		}
		return false
	}
}
