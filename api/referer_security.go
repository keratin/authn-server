package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/keratin/authn-server/services"
)

func RefererSecurity(domains []string) SecurityHandler {
	// optimize for lookups
	domainMap := make(map[string]bool)
	for _, i := range domains {
		domainMap[strings.ToLower(i)] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url, err := url.Parse(r.Header.Get("Referer"))
			if err != nil {
				panic(err)
			}

			if domainMap[url.Hostname()] {
				h.ServeHTTP(w, r)
			} else {
				WriteJson(w, http.StatusForbidden, ServiceErrors{Errors: services.FieldErrors{{"referer", "is not a trusted host"}}})
			}
		})
	}
}
