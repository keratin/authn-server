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

	return json.Marshal(struct {
		ID                int             `json:"id"`
		Username          string          `json:"username"`
		OauthAccounts     []*OauthAccount `json:"oauth_accounts"`
		LastLoginAt       string          `json:"last_login_at"`
		PasswordChangedAt string          `json:"password_changed_at"`
		Locked            bool            `json:"locked"`
		Deleted           bool            `json:"deleted"`
	}{
		ID:                a.ID,
		Username:          a.Username,
		OauthAccounts:     a.OauthAccounts,
		LastLoginAt:       formattedLastLogin,
		PasswordChangedAt: formattedPasswordChangedAt,
		Locked:            a.Locked,
		Deleted:           a.DeletedAt != nil,
	})
}
