package postgres

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/app/models"
)

type AccountStore struct {
	sqlx.Ext
}

func (db *AccountStore) Find(id int) (*models.Account, error) {
	account := models.Account{}
	err := sqlx.Get(db, &account, "SELECT * FROM accounts WHERE id = $1", id)
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
	err := sqlx.Get(db, &account, "SELECT * FROM accounts WHERE username = $1 AND deleted_at IS NULL", u)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &account, nil
}

func (db *AccountStore) FindByOauthAccount(provider string, providerID string) (*models.Account, error) {
	account := models.Account{}
	err := sqlx.Get(db, &account, "SELECT a.* FROM accounts a INNER JOIN oauth_accounts oa ON a.id = oa.account_id WHERE oa.provider = $1 AND oa.provider_id = $2", provider, providerID)
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

	result, err := sqlx.NamedQuery(db,
		`INSERT INTO accounts (
			username,
			password,
			locked,
			require_new_password,
			password_changed_at,
			created_at,
			updated_at
		)
		VALUES (:username, :password, :locked, :require_new_password, :password_changed_at, :created_at, :updated_at)
		RETURNING id`,
		account,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	result.Next()
	var id int64
	err = result.Scan(&id)
	if err != nil {
		return nil, err
	}
	account.ID = int(id)

	return account, nil
}

func (db *AccountStore) AddOauthAccount(accountID int, provider, providerID, email, accessToken string) error {
	now := time.Now()

	_, err := sqlx.NamedExec(db, `
        INSERT INTO oauth_accounts (account_id, provider, provider_id, email, access_token, created_at, updated_at)
        VALUES (:account_id, :provider, :provider_id, :email, :access_token, :created_at, :updated_at)
    `, map[string]interface{}{
		"email":        email,
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
	err := sqlx.Select(db, &accounts, `SELECT * FROM oauth_accounts WHERE account_id = $1`, accountID)
	return accounts, err
}

func (db *AccountStore) UpdateOauthAccount(accountId int, provider, email string) (bool, error) {
	result, err := db.Exec("UPDATE oauth_accounts SET email = $1 WHERE account_id = $2 AND provider = $3", email, accountId, provider)
	if err != nil {
		return false, err
	}

	return ok(result, err)
}

func (db *AccountStore) DeleteOauthAccount(accountId int, provider string) (bool, error) {
	result, err := db.Exec("DELETE FROM oauth_accounts WHERE account_id = $1 AND provider = $2", accountId, provider)
	if err != nil {
		return false, err
	}

	return ok(result, err)
}

func (db *AccountStore) Archive(id int) (bool, error) {
	_, err := db.Exec("DELETE FROM oauth_accounts WHERE account_id = $1", id)
	if err != nil {
		return false, err
	}
	result, err := db.Exec(`
		UPDATE accounts
		SET
			username = CONCAT('@', MD5(RANDOM()::TEXT)),
			password = $1,
			deleted_at = $2
		WHERE id = $3`, "", time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) Lock(id int) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET locked = $1, updated_at = $2 WHERE id = $3", true, time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) Unlock(id int) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET locked = $1, updated_at = $2 WHERE id = $3", false, time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) RequireNewPassword(id int) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET require_new_password = $1, updated_at = $2, totp_secret = null WHERE id = $3", true, time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) SetPassword(id int, p []byte) (bool, error) {
	result, err := db.Exec(`
		UPDATE accounts
		SET
			password = $1,
			require_new_password = $2,
			password_changed_at = $3,
			updated_at = $4
		WHERE
			id = $5`, p, false, time.Now(), time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) UpdateUsername(id int, u string) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET username = $1, updated_at = $2 WHERE id = $3", u, time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) SetLastLogin(id int) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET last_login_at = $1 WHERE id = $2", time.Now(), id)
	return ok(result, err)
}

func (db *AccountStore) SetTOTPSecret(id int, secret []byte) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET totp_secret = $1 WHERE id = $2", secret, id)
	return ok(result, err)
}

func (db *AccountStore) DeleteTOTPSecret(id int) (bool, error) {
	result, err := db.Exec("UPDATE accounts SET totp_secret = NULL WHERE id = $1", id)
	return ok(result, err)
}

func ok(result sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}
