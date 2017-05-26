package handlers

import "github.com/keratin/authn/data/sqlite3"

type App struct {
	Db sqlite3.AccountStore
}
