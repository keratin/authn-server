package mock

import "github.com/keratin/authn/models"

type RefreshTokenStore struct {
}

func (s *RefreshTokenStore) Create(account_id int) (models.RefreshToken, error) {
	return models.RefreshToken("RefreshToken"), nil
}

func (s *RefreshTokenStore) Find(t models.RefreshToken) (int, error) {
	return 0, nil
}

func (s *RefreshTokenStore) Touch(t models.RefreshToken, account_id int) error {
	return nil
}

func (s *RefreshTokenStore) FindAll(account_id int) ([]models.RefreshToken, error) {
	return []models.RefreshToken{}, nil
}

func (s *RefreshTokenStore) Revoke(t models.RefreshToken) error {
	return nil
}
