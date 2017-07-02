package sqlite3

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(env string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", fmt.Sprintf("./%v.db", env))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TempDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	err = MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
