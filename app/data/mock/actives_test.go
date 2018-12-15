package mock_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/data/testers"
)

func TestActives(t *testing.T) {
	for _, tester := range testers.ActivesTesters {
		mStore := mock.NewActives()
		tester(t, mStore)
	}
}
