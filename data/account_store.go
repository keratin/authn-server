package data

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/data/mysql"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/keratin/authn-server/models"
)

type AccountStore interface {
	Create(u string, p []byte) (*models.Account, error)
	Find(id int) (*models.Account, error)
	FindByUsername(u string) (*models.Account, error)
	Archive(id int) error
	Lock(id int) error
	Unlock(id int) error
	RequireNewPassword(id int) error
	SetPassword(id int, p []byte) error
	UpdateUsername(id int, u string) error
}

func NewAccountStore(db *sqlx.DB) (AccountStore, error) {
	switch db.DriverName() {
	case "sqlite3":
		return &sqlite3.AccountStore{DB: db}, nil
	case "mysql":
		return &mysql.AccountStore{DB: db}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %v", db.DriverName())
	}
}
