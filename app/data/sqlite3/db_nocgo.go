// +build nocgo

// This stub package allows authn-server to be build without CGO, i.e. with CGO_ENABLED=0. This has numerous advantages
// including on toolchains where static linking is impossible. In order to use it run with `go build -tags nocgo`.

package sqlite3

import (
	"github.com/jmoiron/sqlx"
)

type Error struct {
	ExtendedCode int
}

func (Error) Error() string {
	panic("Cannot use sqlite3 Error because building with nocgo build tag")
}

var ErrConstraintUnique int

func NewDB(env string) (*sqlx.DB, error) {
	panic("Cannot use sqlite3 DB because building with nocgo build tag")
}

func TestDB() (*sqlx.DB, error) {
	panic("Cannot use sqlite3 DB because building with nocgo build tag")
}
