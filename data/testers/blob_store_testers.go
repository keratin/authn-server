package testers

import (
	"testing"

	"github.com/keratin/authn-server/data"
	"github.com/stretchr/testify/assert"
)

var BlobStoreTesters = []func(*testing.T, data.BlobStore){
	testReadWrite,
	testWriteLock,
}

func testReadWrite(t *testing.T, bs data.BlobStore) {
	blob, err := bs.Read("unknown")
	assert.NoError(t, err)
	assert.Empty(t, blob)

	err = bs.Write("blob", []byte("val"))
	assert.NoError(t, err)

	blob, err = bs.Read("blob")
	assert.NoError(t, err)
	assert.Equal(t, "val", string(blob))
}

func testWriteLock(t *testing.T, bs data.BlobStore) {
	ok, err := bs.WLock("lockedKey")
	assert.NoError(t, err)
	assert.True(t, ok)

	// can't re-acquire, even with the same connection
	ok, err = bs.WLock("lockedKey")
	assert.NoError(t, err)
	assert.False(t, ok)

	// write lock does not create findable blob
	blob, err := bs.Read("lockedKey")
	assert.NoError(t, err)
	assert.Empty(t, blob)

	// can't lock an existing key
	err = bs.Write("existing", []byte("val"))
	assert.NoError(t, err)
	ok, err = bs.WLock("existing")
	assert.NoError(t, err)
	assert.False(t, ok)
}
