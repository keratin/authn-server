package route

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuthSecurity is a SecurityHandler that relies on HTTP Basic Auth. It takes precaution to
// ensure that verifying credentials is a constant time operation, and will not even allow a timing
// attack to confirm if the guessed username is correct without a correct password.
func BasicAuthSecurity(username string, password string, realm string) SecurityHandler {

	// SECURITY: ensure that both ConstantTimeCompare operations are run, so that a
	// timing attack may not verify a correct username without a correct password.
	// this is unable to hide the correct lengths of either, however.
	match := func(u string, p string) bool {
		usernameMatch := subtle.ConstantTimeCompare([]byte(u), []byte(username))
		passwordMatch := subtle.ConstantTimeCompare([]byte(p), []byte(password))

		return (usernameMatch & passwordMatch) == 1
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok || !match(user, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				w.WriteHeader(401)
				_, _ = w.Write([]byte("Unauthorized.\n"))
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
