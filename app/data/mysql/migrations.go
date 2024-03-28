package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	migrations := []func(db *sqlx.DB) error{
		createAccounts,
		createOauthAccounts,
		createAccountLastLoginAtField,
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

func createAccounts(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id INT(11) NOT NULL AUTO_INCREMENT,
            username VARCHAR(255) DEFAULT NULL,
            password VARCHAR(255) DEFAULT NULL,
            locked TINYINT(1) NOT NULL DEFAULT '0',
            require_new_password TINYINT(1) NOT NULL DEFAULT '0',
            password_changed_at DATETIME DEFAULT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            deleted_at DATETIME DEFAULT NULL,
            PRIMARY KEY (id),
            UNIQUE KEY index_accounts_on_username (username)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8
    `)
	return err
}

func createOauthAccounts(db *sqlx.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS oauth_accounts (
            id INT(11) NOT NULL AUTO_INCREMENT,
            account_id INT(11) NOT NULL,
            provider VARCHAR(255) NOT NULL,
            provider_id VARCHAR(255) NOT NULL,
            access_token VARCHAR(255) NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            PRIMARY KEY (id),
            UNIQUE KEY index_oauth_accounts_by_identity (provider_id, provider),
            UNIQUE KEY index_oauth_accounts_by_account_id (account_id, provider)
        )
    `)
	return err
}

func addOauthAccountEmail(db *sqlx.DB) error {
	_, err := db.Exec(`
		ALTER TABLE oauth_accounts ADD COLUMN email VARCHAR(255) DEFAULT NULL;
    `)
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1060 { // 1060 = Duplicate column name
			err = nil
		}
	}
	return err
}

func createAccountLastLoginAtField(db *sqlx.DB) error {
	_, err := db.Exec(`
        ALTER TABLE accounts ADD last_login_at DATETIME DEFAULT NULL
    `)
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1060 { // 1060 = Duplicate column name
			err = nil
		}
	}
	return err
}

func createAccountTOTPFields(db *sqlx.DB) error {
	_, err := db.Exec(`
        ALTER TABLE accounts ADD totp_secret VARCHAR(255) DEFAULT NULL
    `)
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1060 { // 1060 = Duplicate column name
			err = nil
		}
	}
	return err
}
