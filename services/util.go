package services

import (
	"fmt"
	"regexp"
	"strings"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"
var ErrFormatInvalid = "FORMAT_INVALID"
var ErrInsecure = "INSECURE"
var ErrFailed = "FAILED"
var ErrLocked = "LOCKED"
var ErrExpired = "EXPIRED"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Field, e.Message)
}

// worried about an imperfect regex? see: http://www.regular-expressions.info/email.html
var emailPattern = regexp.MustCompile(`(?i)\A[A-Z0-9._%+-]*@(?:[A-Z0-9-]*\.)*[A-Z]*\z`)

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
