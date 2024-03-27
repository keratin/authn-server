package models

import "time"

type OauthAccount struct {
	ID          int
	AccountID   int `db:"account_id"`
	Provider    string
	ProviderID  string    `db:"provider_id"`
	Email       string    `db:"email"`
	AccessToken string    `db:"access_token"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
