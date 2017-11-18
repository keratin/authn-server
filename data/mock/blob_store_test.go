package mock_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/testers"
)

func TestBlobStore(t *testing.T) {
	for _, tester := range testers.BlobStoreTesters {
		store := mock.NewBlobStore(time.Second, time.Second)
		tester(t, store)
	}
}
