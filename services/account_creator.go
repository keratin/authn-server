package services

import (
	"github.com/keratin/authn/data"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func AccountCreator(db data.DB, username string, password string) (*data.Account, []Error) {
	errors := make([]Error, 0, 2)

	if username == "" {
		errors = append(errors, Error{Field: "username", Message: ErrMissing})
	}
	if password == "" {
		errors = append(errors, Error{Field: "password", Message: ErrMissing})
	}

	if len(errors) > 0 {
		return nil, errors
	}

	acc, err := db.Create(username, password)

	if err != nil {
		switch i := err.(type) {
		case sqlite3.Error:
			if i.ExtendedCode == sqlite3.ErrConstraintUnique {
				errors = append(errors, Error{Field: "username", Message: ErrTaken})
				return nil, errors
			} else {
				panic(err)
			}
		default:
			panic(err)
		}
	}

	return acc, nil
}
