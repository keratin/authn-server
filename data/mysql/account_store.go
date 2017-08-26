package mysql

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
	return &account, nil
}

func (db *AccountStore) FindByUsername(u string) (*models.Account, error) {
	account := models.Account{}
	err := db.Get(&account, "SELECT * FROM accounts WHERE username = ?", u)
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
	account.Id = int(id)

	return account, nil
}

func (db *AccountStore) Archive(id int) error {
	_, err := db.Exec("UPDATE accounts SET username = ?, password = ?, deleted_at = ? WHERE id = ?", "", "", time.Now(), id)
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
