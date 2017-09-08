package mock

import (
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

type accountStore struct {
	accountsById map[int]*models.Account
	idByUsername map[string]int
}

func NewAccountStore() *accountStore {
	return &accountStore{
		accountsById: make(map[int]*models.Account),
		idByUsername: make(map[string]int),
	}
}

func (s *accountStore) Find(id int) (*models.Account, error) {
	if s.accountsById[id] != nil {
		return dupAccount(*s.accountsById[id]), nil
	}

	return nil, nil
}

func (s *accountStore) FindByUsername(u string) (*models.Account, error) {
	id := s.idByUsername[u]
	if id == 0 {
		return nil, nil
	}

	return dupAccount(*s.accountsById[id]), nil
}

func (s *accountStore) Create(u string, p []byte) (*models.Account, error) {
	if s.idByUsername[u] != 0 {
		return nil, Error{ErrNotUnique}
	}

	now := time.Now()
	acc := models.Account{
		Id:                len(s.accountsById) + 1,
		Username:          u,
		Password:          p,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	s.accountsById[acc.Id] = &acc
	s.idByUsername[acc.Username] = acc.Id
	return dupAccount(acc), nil
}

func (s *accountStore) Archive(id int) error {
	account := s.accountsById[id]
	if account != nil {
		now := time.Now()
		account.Username = ""
		account.Password = []byte("")
		account.DeletedAt = &now
	}
	return nil
}

func (s *accountStore) Lock(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.Locked = true
		account.UpdatedAt = time.Now()
	}
	return nil
}

func (s *accountStore) Unlock(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.Locked = false
		account.UpdatedAt = time.Now()
	}
	return nil
}

func (s *accountStore) RequireNewPassword(id int) error {
	account := s.accountsById[id]
	if account != nil {
		account.RequireNewPassword = true
		account.UpdatedAt = time.Now()
	}
	return nil
}

func (s *accountStore) SetPassword(id int, p []byte) error {
	account := s.accountsById[id]
	if account != nil {
		now := time.Now()
		account.Password = p
		account.RequireNewPassword = false
		account.PasswordChangedAt = now
		account.UpdatedAt = now
	}
	return nil
}

// i think this works? i want to avoid accidentally giving callers the ability
// to reach into the memory map and modify things or see changes without relying
// on the store api.
func dupAccount(acct models.Account) *models.Account {
	return &acct
}
