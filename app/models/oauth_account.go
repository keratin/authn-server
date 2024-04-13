package models

import (
	"encoding/json"
	"time"
)

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

func (o OauthAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Provider   string `json:"provider"`
		ProviderID string `json:"provider_account_id"`
		Email      string `json:"email"`
	}{
		Provider:   o.Provider,
		ProviderID: o.ProviderID,
		Email:      o.Email,
	})
}
