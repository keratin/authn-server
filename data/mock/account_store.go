package mock

import (
	"database/sql"
	"fmt"
	"time"

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
		return dupAccount(*s.accountsById[id]), nil
	}
}

func (s *AccountStore) Create(u string, p []byte) (*models.Account, error) {
	if s.idByUsername[u] != 0 {
		return nil, Error{ErrNotUnique}
	}

	now := time.Now()
	acc := models.Account{
		Id:        len(s.accountsById) + 1,
		Username:  u,
		Password:  p,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.accountsById[acc.Id] = &acc
	s.idByUsername[acc.Username] = acc.Id
	return dupAccount(acc), nil
}

func (s *AccountStore) Archive(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.Username = ""
		account.Password = []byte{}
		account.DeletedAt = time.Now()
	}
	return nil
}

func (s *AccountStore) Lock(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.Locked = true
		account.UpdatedAt = time.Now()
	}
	return nil
}

func (s *AccountStore) Unlock(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.Locked = false
		account.UpdatedAt = time.Now()
	}
	return nil
}

func (s *AccountStore) RequireNewPassword(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.RequireNewPassword = true
		account.UpdatedAt = time.Now()
	}
	return nil
}

// i think this works? i want to avoid accidentally giving callers the ability
// to reach into the memory map and modify things or see changes without relying
// on the store api.
func dupAccount(acct models.Account) *models.Account {
	return &acct
}
