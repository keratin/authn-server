package services

import (
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
)

func SessionEnder(
	refreshTokenStore data.RefreshTokenStore,
	existingToken *models.RefreshToken,
) (err error) {
	if existingToken != nil {
		return refreshTokenStore.Revoke(*existingToken)
	}
	return nil
}
