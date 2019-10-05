package services

import (
	"regexp"
	"strings"
)

// worried about an imperfect regex? see: http://www.regular-expressions.info/email.html
// NOTE: the {1,125} limit on subdomains has been abbreviated and handled by a length check on the
// string to work around go's parser limitations on nested repetitions.
var emailPattern = regexp.MustCompile(`(?i)\A[A-Z0-9._%+-]{1,64}@(?:[A-Z0-9-]{1,63}\.){1,}[A-Z]{2,63}\z`)

func isEmail(s string) bool {
	// SECURITY: the len() check prevents a regex ddos via overly large usernames
	return len(s) < 255 && emailPattern.Match([]byte(s))
}

func hasDomain(email string, domains []string) bool {
	pieces := strings.SplitN(email, "@", 2)
	for _, domain := range domains {
		if domain == pieces[1] {
			return true
		}
	}
	return false
}
