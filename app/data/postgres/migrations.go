package postgres

import "github.com/jmoiron/sqlx"

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	migrations := []func(db *sqlx.DB) error{
		migrateAccounts,
		createOauthAccounts,
		createAccountLastLoginAtField,
		createAccountTOTPFields,
	}
	for _, m := range migrations {
		if err := m(db); err != nil {
			return err
		}
	}
	return nil
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

func createOauthAccounts(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS oauth_accounts (
            id SERIAL PRIMARY KEY,
            account_id INTEGER NOT NULL,
            provider TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            access_token TEXT NOT NULL,
            created_at timestamptz NOT NULL,
            updated_at timestamptz NOT NULL,
            UNIQUE (provider_id, provider),
            UNIQUE (account_id, provider)
        )
    `)
	return err
}

func createAccountLastLoginAtField(db *sqlx.DB) error {
	_, err := db.Exec(`
        ALTER TABLE accounts ADD COLUMN IF NOT EXISTS last_login_at timestamptz DEFAULT NULL
    `)
	return err
}

func createAccountTOTPFields(db *sqlx.DB) error {
	if _, err := db.Exec(`
        ALTER TABLE accounts ADD COLUMN IF NOT EXISTS totp_enabled boolean DEFAULT false
    `); err != nil {
		return err
	}
	if _, err := db.Exec(`
        ALTER TABLE accounts ADD COLUMN IF NOT EXISTS totp_secret TEXT DEFAULT NULL
    `); err != nil {
		return err
	}
	return nil
}
