package mock

import (
	"fmt"

	"github.com/keratin/authn/models"
)

type RefreshTokenStore struct {
	tokensByAccount map[int][]models.RefreshToken
	accountByToken  map[models.RefreshToken]int
}

func NewRefreshTokenStore() *RefreshTokenStore {
	return &RefreshTokenStore{
		tokensByAccount: make(map[int][]models.RefreshToken),
		accountByToken:  make(map[models.RefreshToken]int),
	}
}

func (s *RefreshTokenStore) Create(account_id int) (models.RefreshToken, error) {
	token := models.RefreshToken(fmt.Sprintf("RefreshToken:%v", account_id))
	s.tokensByAccount[account_id] = append(s.tokensByAccount[account_id], token)
	s.accountByToken[token] = account_id
	return token, nil
}

func (s *RefreshTokenStore) Find(t models.RefreshToken) (int, error) {
	return s.accountByToken[t], nil
}

func (s *RefreshTokenStore) Touch(t models.RefreshToken, account_id int) error {
	return nil
}

func (s *RefreshTokenStore) FindAll(account_id int) ([]models.RefreshToken, error) {
	return s.tokensByAccount[account_id], nil
}

func (s *RefreshTokenStore) Revoke(t models.RefreshToken) error {
	account_id := s.accountByToken[t]
	if account_id != 0 {
		delete(s.accountByToken, t)
		s.tokensByAccount[account_id] = without(t, s.tokensByAccount[account_id])
	}
	return nil
}

func without(needle models.RefreshToken, haystack []models.RefreshToken) []models.RefreshToken {
	for idx, elem := range haystack {
		if elem == needle {
			return append(haystack[:idx], haystack[idx+1:]...)
		}
	}
	return haystack
}
