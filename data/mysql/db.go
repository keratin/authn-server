package mysql

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewDB(url *url.URL) (*sqlx.DB, error) {
	cfg := cfgFromURL(url)
	return sqlx.Connect("mysql", cfg.FormatDSN())
}

// TODO: move to _test
func TestDB() (*sqlx.DB, error) {
	str, ok := os.LookupEnv("TEST_MYSQL_URL")
	if !ok {
		return nil, fmt.Errorf("set TEST_MYSQL_URL for MySQL tests")
	}
	url, err := url.Parse(str)

	err = ensureDB(cfgFromURL(url))
	if err != nil {
		return nil, errors.Wrap(err, "ensureDB")
	}

	db, err := NewDB(url)
	if err != nil {
		return nil, errors.Wrap(err, "NewDB")
	}

	err = MigrateDB(db)
	if err != nil {
		return nil, errors.Wrap(err, "MigrateDB")
	}

	return db, nil
}

func cfgFromURL(url *url.URL) *mysql.Config {
	cfg := mysql.Config{
		Addr:   url.Host,
		DBName: strings.Replace(url.Path, "/", "", 1),
		Loc:    time.UTC,
		Net:    "tcp",
		Params: map[string]string{"parseTime": "true"},
	}
	if url.Port() == "" {
		cfg.Addr = cfg.Addr + ":3306"
	}
	if url.User != nil {
		cfg.User = url.User.Username()
		if pwd, ok := url.User.Password(); ok {
			cfg.Passwd = pwd
		}
	}
	return &cfg
}

func ensureDB(cfg *mysql.Config) error {
	dbName := cfg.DBName
	cfg.DBName = ""

	db, err := sqlx.Connect("mysql", cfg.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "Connect")
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		return errors.Wrap(err, "CREATE DATABASE")
	}

	return nil
}
