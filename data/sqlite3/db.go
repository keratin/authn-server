package sqlite3

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	// load sqlite3 library with side effects
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(env string) (*sqlx.DB, error) {
	// https://github.com/mattn/go-sqlite3/issues/274#issuecomment-232942571
	// enable a busy timeout for concurrent load. keep it short. the busy timeout can be harmful
	// under sustained load, but helpful during short bursts.
	
	// this block used to keep backward compatibility 
	if !strings.Contains(env, ".") {
		env = "./"+ env +".db"
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
