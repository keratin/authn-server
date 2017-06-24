package sqlite3

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(env string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("./%v.db", env))
	if err != nil {
		return nil, err
	}

	err = migrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TempDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}

	err = migrateDB(db)
	if err != nil {
		panic(err)
	}

	return db, nil
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id INTEGER PRIMARY KEY,
            username TEXT CONSTRAINT uniq UNIQUE,
            password TEXT,
            locked INTEGER,
            require_new_password INTEGER,
            password_changed_at INTEGER,
            created_at INTEGER,
            updated_at INTEGER,
            deleted_at INTEGER
        )
    `)
	if err != nil {
		return err
	}
	return nil
}
