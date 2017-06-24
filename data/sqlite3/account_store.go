package sqlite3

import (
	"database/sql"
	"time"

	"github.com/keratin/authn-server/models"
)

type AccountStore struct {
	*sql.DB
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
