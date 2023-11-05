package sqlite3

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"

	// load sqlite3 library with side effects
	_ "modernc.org/sqlite"
)

func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

func NewDB(env string) (*sqlx.DB, error) {
	// https://modernc.org/sqlite/issues/274#issuecomment-232942571
	// enable a busy timeout for concurrent load. keep it short. the busy timeout can be harmful
	// under sustained load, but helpful during short bursts.

	// this block used to keep backward compatibility
	if !strings.Contains(env, ".") {
		env = "./" + env + ".db"
	}

	return sqlx.Connect("sqlite3", fmt.Sprintf("%v?cache=shared&_busy_timeout=200", env))
}

func TestDB() (*sqlx.DB, error) {
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
