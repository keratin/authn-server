package sqlite3

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/models"
)

type AccountStore struct {
	*sqlx.DB
}

func (db *AccountStore) Find(id int) (*models.Account, error) {
	account := models.Account{}
	err := db.Get(&account, "SELECT * FROM accounts WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if account.DeletedAt != nil {
		account.Username = ""
	}
	return &account, nil
}

func (db *AccountStore) FindByUsername(u string) (*models.Account, error) {
	account := models.Account{}
	err := db.Get(&account, "SELECT * FROM accounts WHERE username = ? AND deleted_at IS NULL", u)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &account, nil
}

func (db *AccountStore) FindByOauthAccount(provider string, providerID string) (*models.Account, error) {
	account := models.Account{}
	err := db.Get(&account, "SELECT a.* FROM accounts a INNER JOIN oauth_accounts oa ON a.id = oa.account_id WHERE oa.provider = ? AND oa.provider_id = ?", provider, providerID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &account, nil
}

func (db *AccountStore) Create(u string, p []byte) (*models.Account, error) {
	now := time.Now()

	account := &models.Account{
		Username:          u,
		Password:          p,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	result, err := db.NamedExec(
		"INSERT INTO accounts (username, password, locked, require_new_password, password_changed_at, created_at, updated_at) VALUES (:username, :password, :locked, :require_new_password, :password_changed_at, :created_at, :updated_at)",
		account,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	account.ID = int(id)

	return account, nil
}

func (db *AccountStore) AddOauthAccount(accountID int, provider string, providerID string, accessToken string) error {
	now := time.Now()

	_, err := db.NamedExec(`
        INSERT INTO oauth_accounts (account_id, provider, provider_id, access_token, created_at, updated_at)
        VALUES (:account_id, :provider, :provider_id, :access_token, :created_at, :updated_at)
    `, map[string]interface{}{
		"account_id":   accountID,
		"provider":     provider,
		"provider_id":  providerID,
		"access_token": accessToken,
		"created_at":   now,
		"updated_at":   now,
	})
	return err
}

func (db *AccountStore) GetOauthAccounts(accountID int) ([]*models.OauthAccount, error) {
	accounts := []*models.OauthAccount{}
	err := db.Select(&accounts, `SELECT * FROM oauth_accounts WHERE account_id = ?`, accountID)
	return accounts, err
}

func (db *AccountStore) Archive(id int) error {
	_, err := db.Exec("UPDATE accounts SET username = '@'||HEX(RANDOMBLOB(16)), password = ?, deleted_at = ? WHERE id = ?", "", time.Now(), id)
	return err
}

func (db *AccountStore) Lock(id int) error {
	_, err := db.Exec("UPDATE accounts SET locked = ?, updated_at = ? WHERE id = ?", true, time.Now(), id)
	return err
}

func (db *AccountStore) Unlock(id int) error {
	_, err := db.Exec("UPDATE accounts SET locked = ?, updated_at = ? WHERE id = ?", false, time.Now(), id)
	return err
}

func (db *AccountStore) RequireNewPassword(id int) error {
	_, err := db.Exec("UPDATE accounts SET require_new_password = ?, updated_at = ? WHERE id = ?", true, time.Now(), id)
	return err
}

func (db *AccountStore) SetPassword(id int, p []byte) error {
	_, err := db.Exec("UPDATE accounts SET password = ?, require_new_password = ?, password_changed_at = ?, updated_at = ? WHERE id = ?", p, false, time.Now(), time.Now(), id)
	return err
}

func (db *AccountStore) UpdateUsername(id int, u string) error {
	_, err := db.Exec("UPDATE accounts SET username = ?, updated_at = ? WHERE id = ?", u, time.Now(), id)
	return err
}
