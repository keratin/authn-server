package sqlite3

import (
	"math/rand"
	"time"
)

func jitter() time.Duration {
	// nolint: gosec
	return time.Duration(rand.Intn(5)) * time.Second
}
