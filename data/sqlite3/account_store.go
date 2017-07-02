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
	result, err := db.Exec("INSERT INTO accounts (username, password, locked, require_new_password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", u, p, false, false, now.Unix(), now.Unix())
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Account{
		Id:        int(id),
		Username:  u,
		Password:  p,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
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
