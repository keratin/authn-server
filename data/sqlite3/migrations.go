package sqlite3

import "database/sql"

// This function is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sql.DB) error {
	return migration1(db)
}

func migration1(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id INTEGER PRIMARY KEY,
            username TEXT CONSTRAINT uniq UNIQUE,
            password TEXT,
            locked INTEGER,
            require_new_password INTEGER,
            password_changed_at INTEGER,
            created_at INTEGER,
            updated_at INTEGER,
            deleted_at INTEGER
        )
    `)
	return err
}
