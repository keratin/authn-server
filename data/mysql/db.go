package mysql

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func NewDB(url *url.URL) (*sqlx.DB, error) {
	cfg := cfgFromUrl(url)
	return sqlx.Connect("mysql", cfg.FormatDSN())
}

func TestDB() (*sqlx.DB, error) {
	if _, err := os.Stat("../.env"); !os.IsNotExist(err) {
		godotenv.Load("../.env")
	}
	if _, err := os.Stat("../../.env"); !os.IsNotExist(err) {
		godotenv.Load("../../.env")
	}

	str, ok := os.LookupEnv("TEST_MYSQL_URL")
	if !ok {
		return nil, fmt.Errorf("set TEST_MYSQL_URL for MySQL tests")
	}
	url, err := url.Parse(str)

	err = ensureDB(cfgFromUrl(url))
	if err != nil {
		return nil, err
	}

	db, err := NewDB(url)
	if err != nil {
		return nil, err
	}

	err = MigrateDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func cfgFromUrl(url *url.URL) *mysql.Config {
	cfg := mysql.Config{
		Addr:   url.Host,
		DBName: strings.Replace(url.Path, "/", "", 1),
		Loc:    time.UTC,
		Net:    "tcp",
		Params: map[string]string{"parseTime": "true"},
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
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		return err
	}

	return nil
}
