package data

import "github.com/keratin/authn-server/models"

type RefreshTokenStore interface {
	// Generates and persists a token for the given accountId.
	Create(accountId int) (models.RefreshToken, error)

	// Finds the accountId that owns the token, if the token is registered and unexpired. An empty
	// value indicates that no active token was found.
	Find(t models.RefreshToken) (int, error)

	// Refreshes the lifetime of the token.
	//
	// Technically could operate without accountId, but in the expected contexts the caller should
	// already know the accountId and can save this operation one query by providing it. This seems
	// important since touching can be a high traffic activity.
	Touch(t models.RefreshToken, accountId int) error

	// Returns all tokens that are active for the specified account.
	FindAll(accountId int) ([]models.RefreshToken, error)

	// Revokes the token and removes it from the set of active tokens for the account. Doesn't error
	// if the token is unknown or already revoked.
	Revoke(t models.RefreshToken) error
}
