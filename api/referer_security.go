package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/services"
)

type matchedDomainKey int

func RefererSecurity(domains []config.Domain) SecurityHandler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url, err := url.Parse(r.Header.Get("Referer"))
			if err != nil {
				panic(err)
			}

			for _, d := range domains {
				if d.Matches(url) {
					ctx := r.Context()
					ctx = context.WithValue(ctx, matchedDomainKey(0), &d)

					h.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			WriteJSON(w, http.StatusForbidden, ServiceErrors{Errors: services.FieldErrors{{"referer", "is not a trusted host"}}})
		})
	}
}

func MatchedDomain(r *http.Request) *config.Domain {
	d, ok := r.Context().Value(matchedDomainKey(0)).(*config.Domain)
	if ok {
		return d
	}
	return nil
}
