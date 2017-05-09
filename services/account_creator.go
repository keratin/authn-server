package services

import (
	"github.com/keratin/authn/data"
)

var MISSING = "MISSING"

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
		panic(err)
	}

	return acc, nil
}
