package mock

import (
	"encoding/hex"

	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/app/models"
)

type refreshTokenStore struct {
	tokensByAccount map[int][]models.RefreshToken
	accountByToken  map[models.RefreshToken]int
}

func NewRefreshTokenStore() *refreshTokenStore {
	return &refreshTokenStore{
		tokensByAccount: make(map[int][]models.RefreshToken),
		accountByToken:  make(map[models.RefreshToken]int),
	}
}

func (s *refreshTokenStore) Create(accountID int) (models.RefreshToken, error) {
	binToken, err := lib.GenerateToken()
	if err != nil {
		return "", err
	}
	token := models.RefreshToken(hex.EncodeToString(binToken))
	s.tokensByAccount[accountID] = append(s.tokensByAccount[accountID], token)
	s.accountByToken[token] = accountID
	return token, nil
}

func (s *refreshTokenStore) Find(t models.RefreshToken) (int, error) {
	return s.accountByToken[t], nil
}

func (s *refreshTokenStore) Touch(t models.RefreshToken, accountID int) error {
	return nil
}

func (s *refreshTokenStore) FindAll(accountID int) ([]models.RefreshToken, error) {
	return s.tokensByAccount[accountID], nil
}

func (s *refreshTokenStore) Revoke(t models.RefreshToken) error {
	accountID := s.accountByToken[t]
	if accountID != 0 {
		delete(s.accountByToken, t)
		s.tokensByAccount[accountID] = without(t, s.tokensByAccount[accountID])
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
