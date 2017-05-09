package services_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/keratin/authn/services"
)

func db() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, username TEXT, password TEXT)")
	if err != nil {
		panic(err)
	}

	return db
}

func TestAccountCreatorSuccess(t *testing.T) {
	db := db()
	acc, errs := services.AccountCreator(*db, "userNAME", "PASSword")
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("%v: %v", err.Field, err.Message)
		}
	}

	if acc.Id == nil {
		t.Errorf("expected: %v\ngot: %v", nil, acc.Id)
	}

	if acc.Username != "userNAME" {
		t.Errorf("expected: %v\ngot: %v", "userNAME", acc.Username)
	}
}

func TestAccountCreatorFailure(t *testing.T) {
	db := db()
	acc, errs := services.AccountCreator(*db, "", "")

	expected := make([]services.Error, 0, 2)
	expected = append(expected, services.Error{"username", "MISSING"})
	expected = append(expected, services.Error{"password", "MISSING"})

	if reflect.DeepEqual(errs, expected) {
		t.Errorf("expected: %v\ngot: %v", expected, errs)
	}

	if acc != nil {
		t.Error("unexpected account return")
	}
}
