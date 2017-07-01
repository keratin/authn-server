package sqlite3

import (
	"database/sql"
	"time"

	"github.com/keratin/authn-server/models"
)

type AccountStore struct {
	*sql.DB
}

// If no row is found, the error will be sql.ErrNoRows
func (db *AccountStore) FindByUsername(u string) (*models.Account, error) {
	account := models.Account{}
	err := db.QueryRow("SELECT * FROM accounts WHERE username = ?", u).Scan(&account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (db *AccountStore) Create(u string, p []byte) (*models.Account, error) {
	// TODO: BeginTx with Context!
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result, err := db.Exec("INSERT INTO accounts (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)", u, p, now.Unix(), now.Unix())
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer tx.Commit()

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
	_, err := db.Exec("UPDATE accounts SET username = ?, password = ?, deleted_at = ? WHERE id = ?", nil, nil, time.Now(), id)
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
