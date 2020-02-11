package services

import (
	"github.com/trustelem/zxcvbn"
)

// CalculatePasswordScore uses zxcvbn algorithm to calculate score of a password.
// This function also prevents long password check attacks from happening,
// which would otherwise incur in high CPU usage.
func CalculatePasswordScore(password string) int {
	// SECURITY: only score the first 100 characters of a password. cheap benchmarks on my current
	//           laptop show that latency for 1e3 characters approaches 180ms, and 1e4 characters
	//           consume 54s.
	if len(password) > 100 {
		password = password[:100]
	}

	strength := zxcvbn.PasswordStrength(password, []string{})

	return strength.Score
}
