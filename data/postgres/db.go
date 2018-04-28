package postgres

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	// load pq library with side effects
	_ "github.com/lib/pq"
)

func NewDB(url *url.URL) (*sqlx.DB, error) {
	return sqlx.Connect("postgres", url.String())
}
