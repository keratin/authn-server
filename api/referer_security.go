package api

import (
	"net/http"
	"net/url"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/services"
)

func RefererSecurity(domains []config.Domain) SecurityHandler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url, err := url.Parse(r.Header.Get("Referer"))
			if err != nil {
				panic(err)
			}

			for _, d := range domains {
				if d.Matches(url) {
					h.ServeHTTP(w, r)
					return
				}
			}
			WriteJson(w, http.StatusForbidden, ServiceErrors{Errors: services.FieldErrors{{"referer", "is not a trusted host"}}})
		})
	}
}
