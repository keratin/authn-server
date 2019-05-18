package data_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/data/private"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotatingKeyStore(t *testing.T) {
	ks := data.NewRotatingKeyStore()

	assert.Empty(t, ks.Keys())
	assert.Empty(t, ks.Key())

	k1, err := private.GenerateKey(256)
	require.NoError(t, err)
	ks.Rotate(k1)

	assert.Equal(t, []*private.Key{k1}, ks.Keys())
	assert.Equal(t, k1, ks.Key())

	k2, err := private.GenerateKey(256)
	require.NoError(t, err)
	ks.Rotate(k2)

	assert.Equal(t, []*private.Key{k1, k2}, ks.Keys())
	assert.Equal(t, k2, ks.Key())

	k3, err := private.GenerateKey(256)
	require.NoError(t, err)
	ks.Rotate(k3)

	assert.Equal(t, []*private.Key{k2, k3}, ks.Keys())
	assert.Equal(t, k3, ks.Key())
}
