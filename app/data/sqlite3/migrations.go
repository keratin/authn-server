package sqlite3

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	migrations := []func(db *sqlx.DB) error{
		createAccounts,
		createRefreshTokens,
		createBlobs,
		createOauthAccounts,
		createAccountLastLoginAtField,
		caseInsensitiveUsername,
		createAccountTOTPFields,
		addOauthAccountEmail,
	}
	for _, m := range migrations {
		if err := m(db); err != nil {
			return err
		}
	}
	return nil
}

func isDuplicateError(e error) bool {
	sqliteError, ok := e.(sqlite3.Error)
	return ok && sqliteError.Code == 1 && strings.Contains(sqliteError.Error(), "duplicate column name")
}

func createAccounts(db *sqlx.DB) error {
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

func createRefreshTokens(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            token TEXT NOT NULL CONSTRAINT uniq UNIQUE,
            account_id INTEGER NOT NULL,
            expires_at DATETIME NOT NULL
        )
    `)
	return err
}

func createBlobs(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS blobs (
            name TEXT NOT NULL CONSTRAINT uniq UNIQUE,
            blob BLOB NOT NULL,
            expires_at DATETIME NOT NULL
        )
    `)
	return err
}

func createOauthAccounts(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS oauth_accounts (
            id INTEGER PRIMARY KEY,
            account_id INTEGER,
            provider TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            access_token TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            UNIQUE(provider_id, provider),
            UNIQUE(account_id, provider)
        )
    `)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
        CREATE INDEX IF NOT EXISTS oauth_accounts_by_account_id ON oauth_accounts (account_id)
    `)
	return err
}

func addOauthAccountEmail(db *sqlx.DB) error {
	_, err := db.Exec(`
		ALTER TABLE oauth_accounts ADD COLUMN email VARCHAR(255) DEFAULT NULL;
    `)
	if isDuplicateError(err) {
		return nil
	}
	return err
}

func createAccountLastLoginAtField(db *sqlx.DB) error {
	_, err := db.Exec(`
        ALTER TABLE accounts ADD last_login_at DATETIME
    `)
	if isDuplicateError(err) {
		return nil
	}
	return err
}

// caseInsensitiveUsername will migrate the accounts table to use COLLATE NOCASE on username.
// this will fail if the current accounts table has existing usernames that are equal after
// the operation.
func caseInsensitiveUsername(db *sqlx.DB) error {
	_, err := db.Exec(`
        BEGIN TRANSACTION;

        ALTER TABLE accounts RENAME TO accounts_old;

        CREATE TABLE accounts (
            id INTEGER PRIMARY KEY,
            username TEXT NOT NULL COLLATE NOCASE CONSTRAINT uniq UNIQUE,
            password TEXT NOT NULL,
            locked BOOLEAN NOT NULL,
            require_new_password BOOLEAN NOT NULL,
            password_changed_at DATETIME NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            deleted_at DATETIME,
            last_login_at DATETIME
        );

        INSERT INTO accounts(id, username, password, locked, require_new_password, password_changed_at, created_at, updated_at, deleted_at, last_login_at)
        SELECT id, username, password, locked, require_new_password, password_changed_at, created_at, updated_at, deleted_at, last_login_at
        FROM accounts_old;

        DROP TABLE accounts_old;

        COMMIT;
    `)
	return err
}

func createAccountTOTPFields(db *sqlx.DB) error {
	_, err := db.Exec(`
        ALTER TABLE accounts ADD totp_secret VARCHAR(255) DEFAULT NULL
    `)
	if isDuplicateError(err) {
		return nil
	}
	return err
}
