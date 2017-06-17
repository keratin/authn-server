package mock

import (
	"fmt"

	"github.com/keratin/authn/models"
)

type Error struct {
	Code int
}

func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Code)
}

const ErrNotUnique = iota

type AccountStore struct {
	OnCreate func(u string, p []byte) (*models.Account, error)
}

func (s *AccountStore) Create(u string, p []byte) (*models.Account, error) {
	return s.OnCreate(u, p)
}
