package postgres

import "github.com/jmoiron/sqlx"

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	return migrateAccounts(db)
}
func migrateAccounts(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id SERIAL PRIMARY KEY,
            username TEXT UNIQUE DEFAULT NULL,
            password TEXT DEFAULT NULL,
            locked boolean NOT NULL DEFAULT false,
            require_new_password boolean NOT NULL DEFAULT false,
            password_changed_at timestamptz DEFAULT NULL,
            created_at timestamptz NOT NULL,
            updated_at timestamptz NOT NULL,
            deleted_at timestamptz DEFAULT NULL
        )
    `)
	return err
}
