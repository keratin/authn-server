package data

import (
	"fmt"

	"github.com/keratin/authn-server/data/postgres"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/data/mysql"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/keratin/authn-server/app/models"
)

type AccountStore interface {
	Create(u string, p []byte) (*models.Account, error)
	Find(id int) (*models.Account, error)
	FindByUsername(u string) (*models.Account, error)
	FindByOauthAccount(p string, pid string) (*models.Account, error)
	AddOauthAccount(id int, p string, pid string, tok string) error
	GetOauthAccounts(id int) ([]*models.OauthAccount, error)
	Archive(id int) (bool, error)
	Lock(id int) (bool, error)
	Unlock(id int) (bool, error)
	RequireNewPassword(id int) (bool, error)
	SetPassword(id int, p []byte) (bool, error)
	UpdateUsername(id int, u string) (bool, error)
	SetLastLogin(id int) (bool, error)
}

func NewAccountStore(db *sqlx.DB) (AccountStore, error) {
	switch db.DriverName() {
	case "sqlite3":
		return &sqlite3.AccountStore{DB: db}, nil
	case "mysql":
		return &mysql.AccountStore{DB: db}, nil
	case "postgres":
		return &postgres.AccountStore{DB: db}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %v", db.DriverName())
	}
}
