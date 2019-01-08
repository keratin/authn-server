package mock_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/data/testers"
)

func TestAccountStore(t *testing.T) {
	for _, tester := range testers.AccountStoreTesters {
		store := mock.NewAccountStore()
		tester(t, store)
	}
}
