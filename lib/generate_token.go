package lib

import "crypto/rand"

// GenerateToken returns 128 bits of randmoness. This is more than a UUID v4.
// cf: https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_.28random.29
func GenerateToken() ([]byte, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return []byte{}, nil
	}
	return token, nil
}
