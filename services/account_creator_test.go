package services_test

import (
	"reflect"
	"testing"

	"github.com/keratin/authn/data"
	"github.com/keratin/authn/services"
)

type dict map[string]string

func (dict *dict) toServiceErrors() []services.Error {
	errs := make([]services.Error, 0, len(*dict))
	for field, message := range *dict {
		errs = append(errs, services.Error{field, message})
	}
	return errs
}

func TestAccountCreatorSuccess(t *testing.T) {
	db, err := data.TempDB()
	if err != nil {
		panic(err)
	}

	acc, errs := services.AccountCreator(*db, "userNAME", "PASSword")
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
	db, err := data.TempDB()
	if err != nil {
		panic(err)
	}

	db.Create("existing@test.com", "random")

	var tests = []struct {
		username string
		password string
		errors   dict
	}{
		{"", "", dict{"username": "MISSING", "password": "MISSING"}},
		{"existing@test.com", "PASSword", dict{"username": "TAKEN"}},
	}

	for _, tt := range tests {
		acc, errs := services.AccountCreator(*db, tt.username, tt.password)
		if acc != nil {
			t.Error("unexpected account return")
		}
		expected := tt.errors.toServiceErrors()
		if !reflect.DeepEqual(expected, errs) {
			t.Errorf("\nexpected: %v\ngot: %v", expected, errs)
		}
	}
}
