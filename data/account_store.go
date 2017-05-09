package data

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
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

func NewDB(env string) (*DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("./%v.db", env))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, username TEXT, password TEXT)")
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Create(u string, p string) (*Account, error) {
	// TODO: BeginTx with Context!
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// TODO: bcrypt password
	result, err := db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", u, p)
	if err != nil {
		// TODO: detect and handle uniqueness failures
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
