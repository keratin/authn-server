package sqlite3

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/ops"
	sq3 "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var placeholder = "generating"

type BlobStore struct {
	TTL      time.Duration
	LockTime time.Duration
	DB       sqlx.Ext
}

func (s *BlobStore) Clean(reporter ops.ErrorReporter) {
	go func() {
		for range time.Tick(time.Minute + time.Duration(rand.Intn(5))*time.Second) {
			_, err := s.DB.Exec("DELETE FROM blobs WHERE expires_at < ?", time.Now())
			if err != nil {
				reporter.ReportError(errors.Wrap(err, "BlobStore Clean"))
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (s *BlobStore) Read(name string) ([]byte, error) {
	var blob []byte
	err := s.DB.QueryRowx("SELECT blob FROM blobs WHERE name = ? AND blob != ? AND expires_at > ?", name, placeholder, time.Now()).Scan(&blob)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	return blob, nil
}

func (s *BlobStore) WriteNX(name string, blob []byte) (bool, error) {
	_, err := s.DB.Exec("INSERT INTO blobs (name, blob, expires_at) VALUES (?, ?, ?)", name, blob, time.Now().Add(s.TTL))
	if i, ok := err.(sq3.Error); ok && i.ExtendedCode == sq3.ErrConstraintUnique {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
