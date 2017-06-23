package mock

import (
	"database/sql"
	"fmt"

	"github.com/keratin/authn-server/models"
)

type Error struct {
	Code int
}

func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Code)
}

const ErrNotUnique = iota

type AccountStore struct {
	accountsById map[int]*models.Account
	idByUsername map[string]int
}

func NewAccountStore() *AccountStore {
	return &AccountStore{
		accountsById: make(map[int]*models.Account),
		idByUsername: make(map[string]int),
	}
}

func (s *AccountStore) FindByUsername(u string) (*models.Account, error) {
	id := s.idByUsername[u]
	if id == 0 {
		return nil, sql.ErrNoRows
	} else {
		return s.accountsById[id], nil
	}
}

func (s *AccountStore) Create(u string, p []byte) (*models.Account, error) {
	if s.idByUsername[u] != 0 {
		return nil, Error{ErrNotUnique}
	}

	acc := &models.Account{
		Id:       len(s.accountsById) + 1,
		Username: u,
		Password: p,
	}
	s.accountsById[acc.Id] = acc
	s.idByUsername[acc.Username] = acc.Id
	return acc, nil
}
