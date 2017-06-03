package services_test

import (
	"reflect"
	"testing"

	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data/sqlite3"
	"github.com/keratin/authn/services"
)

func TestAccountCreatorSuccess(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	var tests = []struct {
		config   config.Config
		username string
		password string
	}{
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "userName", "PASSword"},
		{config.Config{UsernameIsEmail: true}, "username@test.com", "PASSword"},
		{config.Config{UsernameIsEmail: true, UsernameDomain: "rightdomain.com"}, "username@rightdomain.com", "PASSword"},
	}

	for _, tt := range tests {
		acc, errs := services.AccountCreator(&store, &tt.config, tt.username, tt.password)
		if len(errs) > 0 {
			for _, err := range errs {
				t.Errorf("%v: %v", err.Field, err.Message)
			}
		}

		if acc != nil && acc.Id == 0 {
			t.Errorf("\nexpected: %v\ngot: %v", nil, acc.Id)
		}

		if acc != nil && acc.Username != tt.username {
			t.Errorf("\nexpected: %v\ngot: %v", tt.username, acc.Username)
		}
	}
}

var pw = []byte("$2a$04$ZOBA8E3nT68/ArE6NDnzfezGWEgM6YrE17PrOtSjT5.U/ZGoxyh7e")

func TestAccountCreatorFailure(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	store.Create("existing@test.com", pw)

	var tests = []struct {
		config   config.Config
		username string
		password string
		errors   []services.Error
	}{
		// username validations
		{config.Config{}, "", "PASSword", []services.Error{{"username", "MISSING"}}},
		{config.Config{}, "  ", "PASSword", []services.Error{{"username", "MISSING"}}},
		{config.Config{}, "existing@test.com", "PASSword", []services.Error{{"username", "TAKEN"}}},
		{config.Config{UsernameIsEmail: true}, "notanemail", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: true, UsernameDomain: "rightdomain.com"}, "email@wrongdomain.com", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		{config.Config{UsernameIsEmail: false, UsernameMinLength: 6}, "short", "PASSword", []services.Error{{"username", "FORMAT_INVALID"}}},
		// password validations
		{config.Config{}, "username", "", []services.Error{{"password", "MISSING"}}},
		{config.Config{PasswordMinComplexity: 2}, "username", "qwerty", []services.Error{{"password", "INSECURE"}}},
	}

	for _, tt := range tests {
		acc, errs := services.AccountCreator(&store, &tt.config, tt.username, tt.password)
		if acc != nil {
			t.Error("unexpected account return")
		}
		if !reflect.DeepEqual(tt.errors, errs) {
			t.Errorf("\nexpected: %v\ngot: %v", tt.errors, errs)
		}
	}
}
