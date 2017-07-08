package api

import (
	"crypto/subtle"
	"net/http"
)

func BasicAuthSecurity(username string, password string, realm string) SecurityHandler {

	// SECURITY: ensure that both ConstantTimeCompare operations are run, so that a
	// timing attack may not verify a correct username without a correct password.
	match := func(u string, p string) bool {
		username_match := subtle.ConstantTimeCompare([]byte(u), []byte(username))
		password_match := subtle.ConstantTimeCompare([]byte(p), []byte(password))

		return username_match == 1 && password_match == 1
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok || !match(user, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				w.WriteHeader(401)
				w.Write([]byte("Unauthorized.\n"))
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
