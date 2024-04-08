package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Account struct {
	ID                 int
	Username           string
	Password           []byte
	Locked             bool
	RequireNewPassword bool           `db:"require_new_password"`
	PasswordChangedAt  time.Time      `db:"password_changed_at"`
	TOTPSecret         sql.NullString `db:"totp_secret"`
	OauthAccounts      []*OauthAccount
	LastLoginAt        *time.Time `db:"last_login_at"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at"`
}

func (a Account) Archived() bool {
	return a.DeletedAt != nil
}

// TOTPEnabled returns true if OTP is enabled on the account
func (a Account) TOTPEnabled() bool {
	if a.TOTPSecret.Valid && a.TOTPSecret.String != "" {
		return true
	}
	return false
}

func (a Account) MarshalJSON() ([]byte, error) {
	formattedLastLogin := ""
	if a.LastLoginAt != nil {
		formattedLastLogin = a.LastLoginAt.Format(time.RFC3339)
	}

	formattedPasswordChangedAt := ""
	if !a.PasswordChangedAt.IsZero() {
		formattedPasswordChangedAt = a.PasswordChangedAt.Format(time.RFC3339)
	}

	return json.Marshal(map[string]interface{}{
		"id":                  a.ID,
		"username":            a.Username,
		"oauth_accounts":      a.OauthAccounts,
		"last_login_at":       formattedLastLogin,
		"password_changed_at": formattedPasswordChangedAt,
		"locked":              a.Locked,
		"deleted":             a.DeletedAt != nil,
	})
}
