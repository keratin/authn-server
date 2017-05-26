package data

import (
	"database/sql"
)

type Account struct {
	Id       int
	Username string
}

type AccountStore interface {
	Create(u string, p string) (*Account, error)
}

type DB struct {
	*sql.DB
}

func (db *DB) Create(u string, p string) (*Account, error) {
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

	account := Account{Id: int(id), Username: u}

	return &account, nil
}
