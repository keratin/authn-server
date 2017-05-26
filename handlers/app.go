package handlers

import (
	"database/sql"

	"github.com/keratin/authn/data/sqlite3"
)

type App struct {
	Db           sql.DB
	AccountStore sqlite3.AccountStore
}
