package mysql

import "github.com/jmoiron/sqlx"

// MigrateDB is committed to doing the work necessary to converge the database
// in a safe, production-grade fashion. This will mean conditional logic as it
// determines which steps have run and which steps must still be run. Given the
// expected final complexity of this project, this is acceptable.
func MigrateDB(db *sqlx.DB) error {
	return migration1(db)
}

func migration1(db *sqlx.DB) error {
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
