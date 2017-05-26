package services

import (
	"github.com/keratin/authn/data"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var MISSING = "MISSING"
var TAKEN = "TAKEN"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func AccountCreator(db data.DB, username string, password string) (*data.Account, []Error) {
	errors := make([]Error, 0, 2)

	if username == "" {
		errors = append(errors, Error{Field: "username", Message: MISSING})
	}
	if password == "" {
		errors = append(errors, Error{Field: "password", Message: MISSING})
	}

	if len(errors) > 0 {
		return nil, errors
	}

	acc, err := db.Create(username, password)

	if err != nil {
		switch i := err.(type) {
		case sqlite3.Error:
			if i.Code == sqlite3.ErrConstraint {
				errors = append(errors, Error{Field: "username", Message: TAKEN})
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
