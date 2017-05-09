package services

import (
	"database/sql"
)

var MISSING = "MISSING"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Account struct {
	Id       int
	Username string
}

func AccountCreator(db sql.DB, username string, password string) (*Account, []Error) {
	errors := make([]Error, 0, 1)

	if username == "" {
		errors = append(errors, Error{Field: "username", Message: MISSING})
	}
	if password == "" {
		errors = append(errors, Error{Field: "password", Message: MISSING})
	}

	if len(errors) > 0 {
		return nil, errors
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	result, err := db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	account := Account{Id: int(id), Username: username}

	tx.Commit()

	return &account, nil
}
