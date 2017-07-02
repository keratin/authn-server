package sqlite3_test

import (
	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/data/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func tempDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	err = sqlite3.MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
