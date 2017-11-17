package sqlite3

import "github.com/jmoiron/sqlx"

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	migrations := []func(db *sqlx.DB) error{
		migration1,
		migration2,
		migration3,
	}
	for _, m := range migrations {
		if err := m(db); err != nil {
			return err
		}
	}
	return nil
}

func migration1(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id INTEGER PRIMARY KEY,
            username TEXT NOT NULL CONSTRAINT uniq UNIQUE,
            password TEXT NOT NULL,
            locked BOOLEAN NOT NULL,
            require_new_password BOOLEAN NOT NULL,
            password_changed_at DATETIME NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            deleted_at DATETIME
        )
    `)
	return err
}

func migration2(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            token TEXT NOT NULL CONSTRAINT uniq UNIQUE,
            account_id INTEGER NOT NULL,
            expires_at DATETIME NOT NULL
        )
    `)
	return err
}

func migration3(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS blobs (
            name TEXT NOT NULL CONSTRAINT uniq UNIQUE,
            blob BLOB NOT NULL,
            expires_at DATETIME NOT NULL
        )
    `)
	return err
}
