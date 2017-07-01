package data

import (
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/sqlite3"
	sq3 "github.com/mattn/go-sqlite3"
)

func NewDB(url *url.URL) (*sqlx.DB, AccountStore, error) {
	switch url.Scheme {
	case "sqlite3":
		db, err := sqlite3.NewDB(url.Path)
		if err != nil {
			return nil, nil, err
		}
		store := sqlite3.AccountStore{db}
		return db, &store, nil
	default:
		return nil, nil, fmt.Errorf("Unsupported database")
	}
}

func MigrateDB(url *url.URL) error {
	switch url.Scheme {
	case "sqlite3":
		db, err := sqlite3.NewDB(url.Path)
		if err != nil {
			return err
		}
		defer db.Close()

		sqlite3.MigrateDB(db)
		return nil
	default:
		return fmt.Errorf("Unsupported database")
	}
}

func IsUniquenessError(err error) bool {
	switch i := err.(type) {
	case sq3.Error:
		return i.ExtendedCode == sq3.ErrConstraintUnique
	case mock.Error:
		return i.Code == mock.ErrNotUnique
	default:
		return false
	}
}
