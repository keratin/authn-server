package data

type RefreshToken string

type RefreshTokenStore interface {
	// Generates and persists a token for the given account_id.
	Create(account_id int) (RefreshToken, error)

	// Finds the account_id that owns the token, if the token is registered and unexpired. An empty
	// value indicates that no active token was found.
	Find(t RefreshToken) (int, error)

	// Refreshes the lifetime of the token.
	//
	// Technically could operate without account_id, but in the expected contexts the caller should
	// already know the account_id and can save this operation one query by providing it. This seems
	// important since touching can be a high traffic activity.
	Touch(t RefreshToken, account_id int) error

	// Returns all tokens that are active for the specified account.
	FindAll(account_id int) ([]RefreshToken, error)

	// Revokes the token and removes it from the set of active tokens for the account. Doesn't error
	// if the token is unknown or already revoked.
	Revoke(t RefreshToken) error
}
