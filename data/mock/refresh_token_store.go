package mock

import (
	"fmt"

	"github.com/keratin/authn-server/models"
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

func (s *RefreshTokenStore) Create(accountId int) (models.RefreshToken, error) {
	token := models.RefreshToken(fmt.Sprintf("RefreshToken:%v", accountId))
	s.tokensByAccount[accountId] = append(s.tokensByAccount[accountId], token)
	s.accountByToken[token] = accountId
	return token, nil
}

func (s *RefreshTokenStore) Find(t models.RefreshToken) (int, error) {
	return s.accountByToken[t], nil
}

func (s *RefreshTokenStore) Touch(t models.RefreshToken, accountId int) error {
	return nil
}

func (s *RefreshTokenStore) FindAll(accountId int) ([]models.RefreshToken, error) {
	return s.tokensByAccount[accountId], nil
}

func (s *RefreshTokenStore) Revoke(t models.RefreshToken) error {
	accountId := s.accountByToken[t]
	if accountId != 0 {
		delete(s.accountByToken, t)
		s.tokensByAccount[accountId] = without(t, s.tokensByAccount[accountId])
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
