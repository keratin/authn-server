package services_test

import (
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tests"
)

func TestCredentialsVerifierSuccess(t *testing.T) {
	username := "myname"
	password := "mysecret"
	bcrypted := []byte("$2a$04$lzQPXlov4RFLxps1uUGq4e4wmVjLYz3WrqQw4bSdfIiJRyo3/fk3C")

	cfg := config.Config{BcryptCost: 4}
	store := mock.NewAccountStore()
	store.Create(username, bcrypted)

	acc, errs := services.CredentialsVerifier(store, &cfg, username, password)
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("%v: %v", err.Field, err.Message)
		}
	} else {
		tests.RefuteEqual(t, 0, acc.Id)
		tests.AssertEqual(t, username, acc.Username)
	}
}

func TestCredentialsVerifierFailure(t *testing.T) {
	password := "mysecret"
	bcrypted := []byte("$2a$04$lzQPXlov4RFLxps1uUGq4e4wmVjLYz3WrqQw4bSdfIiJRyo3/fk3C")

	cfg := config.Config{BcryptCost: 4}
	store := mock.NewAccountStore()
	store.Create("known", bcrypted)
	acc, _ := store.Create("locked", bcrypted)
	acc.Locked = true // this is a reference to the memory store
	acc, _ = store.Create("expired", bcrypted)
	acc.RequireNewPassword = true

	testTable := []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"credentials", "FAILED"}}},
		{"unknown", "unknown", []services.Error{{"credentials", "FAILED"}}},
		{"known", "unknown", []services.Error{{"credentials", "FAILED"}}},
		{"unknown", password, []services.Error{{"credentials", "FAILED"}}},
		{"locked", password, []services.Error{{"account", "LOCKED"}}},
		{"expired", password, []services.Error{{"credentials", "EXPIRED"}}},
	}

	for _, tt := range testTable {
		_, errs := services.CredentialsVerifier(store, &cfg, tt.username, tt.password)
		tests.AssertEqual(t, tt.errors, errs)
	}
}
