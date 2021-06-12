// +build nocgo

// This stub package allows authn-server to be build without CGO, i.e. with CGO_ENABLED=0. This has numerous advantages
// including on toolchains where static linking is impossible. In order to use it run with `go build -tags nocgo`.

package sqlite3

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/keratin/authn-server/ops"
)

var placeholder = "generating"

type BlobStore struct {
	TTL      time.Duration
	LockTime time.Duration
	DB       sqlx.Ext
}

func (s *BlobStore) Clean(reporter ops.ErrorReporter) {
	panic("Cannot use sqlite3 BlobStore because building with nocgo build tag")
}

func (s *BlobStore) Read(name string) ([]byte, error) {
	panic("Cannot use sqlite3 BlobStore because building with nocgo build tag")
}

func (s *BlobStore) WriteNX(name string, blob []byte) (bool, error) {
	panic("Cannot use sqlite3 BlobStore because building with nocgo build tag")
}
