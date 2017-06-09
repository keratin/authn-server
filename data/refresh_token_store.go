package data

type RefreshToken string

type RefreshTokenStore interface {
	Find(t RefreshToken) (int, error)
	Touch(t RefreshToken, account_id int) error
	FindAll(account_id int) ([]RefreshToken, error)
	Create(account_id int) (RefreshToken, error)
	Revoke(t RefreshToken) error
}
