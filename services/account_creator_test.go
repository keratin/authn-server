package services_test

import (
	"reflect"
	"testing"

	"github.com/keratin/authn/data/sqlite3"
	"github.com/keratin/authn/services"
)

func TestAccountCreatorSuccess(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	acc, errs := services.AccountCreator(&store, "userNAME", "PASSword")
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("%v: %v", err.Field, err.Message)
		}
	}

	if acc.Id == 0 {
		t.Errorf("\nexpected: %v\ngot: %v", nil, acc.Id)
	}

	if acc.Username != "userNAME" {
		t.Errorf("\nexpected: %v\ngot: %v", "userNAME", acc.Username)
	}
}

func TestAccountCreatorFailure(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	store.Create("existing@test.com", "random")

	var tests = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"username", "MISSING"}, {"password", "MISSING"}}},
		{"existing@test.com", "PASSword", []services.Error{{"username", "TAKEN"}}},
	}

	for _, tt := range tests {
		acc, errs := services.AccountCreator(&store, tt.username, tt.password)
		if acc != nil {
			t.Error("unexpected account return")
		}
		if !reflect.DeepEqual(tt.errors, errs) {
			t.Errorf("\nexpected: %v\ngot: %v", tt.errors, errs)
		}
	}
}
