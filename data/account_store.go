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

	err = MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func TempDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}

	err = MigrateDB(db)
	if err != nil {
		panic(err)
	}

	return &DB{db}, nil
}

func MigrateDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, username TEXT CONSTRAINT uniq UNIQUE, password TEXT)")
	if err != nil {
		return err
	}
	return nil
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
