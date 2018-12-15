package testers

import (
	"testing"

	"github.com/keratin/authn-server/app/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var BlobStoreTesters = []func(*testing.T, data.BlobStore){
	testRead,
	testWriteNX,
}

func testRead(t *testing.T, bs data.BlobStore) {
	blob, err := bs.Read("unknown")
	assert.NoError(t, err)
	assert.Empty(t, blob)

	ok, err := bs.WriteNX("blob", []byte("val"))
	require.NoError(t, err)
	require.True(t, ok)

	blob, err = bs.Read("blob")
	assert.NoError(t, err)
	assert.Equal(t, "val", string(blob))
}

func testWriteNX(t *testing.T, bs data.BlobStore) {
	set, err := bs.WriteNX("key", []byte("first"))
	assert.NoError(t, err)
	assert.True(t, set)

	set, err = bs.WriteNX("key", []byte("second"))
	assert.NoError(t, err)
	assert.False(t, set)
}
