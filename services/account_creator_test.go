package services_test

import (
	"reflect"
	"testing"

	"github.com/keratin/authn/data"
	"github.com/keratin/authn/services"
)

func TestAccountCreatorSuccess(t *testing.T) {
	db, err := data.NewDB("test")
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
	db, err := data.NewDB("test")
	if err != nil {
		panic(err)
	}

	expected := make([]services.Error, 0, 2)
	expected = append(expected, services.Error{"username", "MISSING"})
	expected = append(expected, services.Error{"password", "MISSING"})

	acc, errs := services.AccountCreator(*db, "", "")
	if !reflect.DeepEqual(errs, expected) {
		t.Errorf("\nexpected: %v\ngot: %v", expected, errs)
	}

	if acc != nil {
		t.Error("unexpected account return")
	}
}
