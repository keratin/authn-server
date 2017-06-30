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

	return db, nil
}

func TempDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	err = MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
