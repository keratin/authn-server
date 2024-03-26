package mock

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keratin/authn-server/app/models"
)

type Error struct {
	Code int
}

func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Code)
}

const ErrNotUnique = iota

type accountStore struct {
	accountsByID      map[int]*models.Account
	idByUsername      map[string]int
	oauthAccountsByID map[int][]*models.OauthAccount
	idByOauthID       map[string]int
	errorOnID         int
}

func WithSetTOTPFailureID(id int) func(s *accountStore) {
	return func(s *accountStore) {
		s.errorOnID = id
	}
}

func NewAccountStore(opts ...func(*accountStore)) *accountStore {
	s := &accountStore{
		accountsByID:      make(map[int]*models.Account),
		oauthAccountsByID: make(map[int][]*models.OauthAccount),
		idByUsername:      make(map[string]int),
		idByOauthID:       make(map[string]int),
		errorOnID:         -1,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *accountStore) Find(id int) (*models.Account, error) {
	if s.accountsByID[id] != nil {
		return dupAccount(*s.accountsByID[id]), nil
	}

	return nil, nil
}

func (s *accountStore) FindByUsername(u string) (*models.Account, error) {
	id := s.idByUsername[strings.ToLower(u)]
	if id == 0 {
		return nil, nil
	}

	return dupAccount(*s.accountsByID[id]), nil
}

func (s *accountStore) FindByOauthAccount(provider string, providerID string) (*models.Account, error) {
	id := s.idByOauthID[provider+"|"+providerID]
	if id == 0 {
		return nil, nil
	}

	return dupAccount(*s.accountsByID[id]), nil
}

func (s *accountStore) Create(u string, p []byte) (*models.Account, error) {
	if s.idByUsername[strings.ToLower(u)] != 0 {
		return nil, Error{ErrNotUnique}
	}

	now := time.Now()
	acc := models.Account{
		ID:                len(s.accountsByID) + 1,
		Username:          u,
		Password:          p,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	s.accountsByID[acc.ID] = &acc
	s.idByUsername[strings.ToLower(acc.Username)] = acc.ID
	return dupAccount(acc), nil
}

func (s *accountStore) AddOauthAccount(accountID int, provider string, providerID string, tok string) error {
	p := provider + "|" + providerID
	if s.idByOauthID[p] != 0 {
		return Error{ErrNotUnique}
	}
	for _, oa := range s.oauthAccountsByID[accountID] {
		if oa.Provider == provider {
			return Error{ErrNotUnique}
		}
	}

	now := time.Now()
	oauthAccount := &models.OauthAccount{
		AccountID:   accountID,
		Provider:    provider,
		ProviderID:  providerID,
		AccessToken: tok,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.oauthAccountsByID[accountID] = append(s.oauthAccountsByID[accountID], oauthAccount)

	s.idByOauthID[p] = accountID
	return nil
}

func (s *accountStore) GetOauthAccounts(accountID int) ([]*models.OauthAccount, error) {
	return s.oauthAccountsByID[accountID], nil
}

func (s *accountStore) DeleteOauthAccount(accountID int, provider string) (bool, error) {
	oauthAccounts := s.oauthAccountsByID[accountID]

	for i, oauthAccount := range oauthAccounts {
		if oauthAccount.Provider == provider {
			s.oauthAccountsByID[accountID] = append(oauthAccounts[:i], oauthAccounts[i+1:]...)
			return true, nil
		}
	}

	return false, nil
}

func (s *accountStore) Archive(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	delete(s.idByUsername, strings.ToLower(account.Username))
	now := time.Now()
	account.Username = ""
	account.Password = []byte("")
	account.DeletedAt = &now

	for _, oauthAccount := range s.oauthAccountsByID[account.ID] {
		delete(s.idByOauthID, oauthAccount.Provider+"|"+oauthAccount.ProviderID)
	}
	delete(s.oauthAccountsByID, account.ID)

	return true, nil
}

func (s *accountStore) Lock(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	account.Locked = true
	account.UpdatedAt = time.Now()
	return true, nil
}

func (s *accountStore) Unlock(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	account.Locked = false
	account.UpdatedAt = time.Now()
	return true, nil
}

func (s *accountStore) RequireNewPassword(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	account.RequireNewPassword = true
	account.UpdatedAt = time.Now()
	account.TOTPSecret = sql.NullString{}

	return true, nil
}

func (s *accountStore) SetPassword(id int, p []byte) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	now := time.Now()
	account.Password = p
	account.RequireNewPassword = false
	account.PasswordChangedAt = now
	account.UpdatedAt = now
	return true, nil
}

func (s *accountStore) UpdateUsername(id int, u string) (bool, error) {
	uNormalized := strings.ToLower(u)
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	if s.idByUsername[uNormalized] != 0 && s.idByUsername[uNormalized] != id {
		return false, Error{ErrNotUnique}
	}

	account.Username = u
	account.UpdatedAt = time.Now()
	s.idByUsername[uNormalized] = account.ID
	return true, nil
}

func (s *accountStore) SetLastLogin(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	now := time.Now()
	account.LastLoginAt = &now
	return true, nil
}

func (s *accountStore) SetTOTPSecret(id int, secret []byte) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}

	// this is weird, but we can return "unaffected" if the secret already exists
	// to approximate the failure mode for testing.
	if account.TOTPSecret.Valid {
		return false, nil
	}

	if account.ID == s.errorOnID {
		return false, fmt.Errorf("rejecting for bad ID: %d", account.ID)
	}

	account.TOTPSecret = sql.NullString{String: string(secret), Valid: true}
	return true, nil
}

func (s *accountStore) DeleteTOTPSecret(id int) (bool, error) {
	account := s.accountsByID[id]
	if account == nil {
		return false, nil
	}
	deleted := false
	if account.TOTPSecret.Valid {
		account.TOTPSecret = sql.NullString{}
		deleted = true
	}
	return deleted, nil
}

// i think this works? i want to avoid accidentally giving callers the ability
// to reach into the memory map and modify things or see changes without relying
// on the store api.
func dupAccount(acct models.Account) *models.Account {
	return &acct
}
