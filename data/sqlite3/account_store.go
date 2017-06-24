package sqlite3

import (
	"database/sql"

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

	result, err := db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", u, p)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer tx.Commit()

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	account := models.Account{Id: int(id), Username: u}

	return &account, nil
}
