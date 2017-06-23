package data

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/keratin/authn/data/mock"
	"github.com/keratin/authn/data/sqlite3"
	sq3 "github.com/mattn/go-sqlite3"
)

func NewDB(url *url.URL) (*sql.DB, AccountStore, error) {
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
