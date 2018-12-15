package testers

import (
	"testing"

	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var RefreshTokenStoreTesters = []func(*testing.T, data.RefreshTokenStore){
	testRefreshTokenFind,
	testRefreshTokenTouch,
	testRefreshTokenFindAll,
	testRefreshTokenCreate,
	testRefreshTokenRevoke,
}

// TODO: find way to test that expired tokens are not found
func testRefreshTokenFind(t *testing.T, store data.RefreshTokenStore) {
	// finding nothing
	id, err := store.Find(models.RefreshToken("a1b2c3"))
	assert.Empty(t, id)
	assert.NoError(t, err)

	// finding something
	id2 := 123
	token, err := store.Create(id2)
	require.NoError(t, err)
	found, err := store.Find(token)
	if assert.NoError(t, err) {
		assert.Equal(t, found, id2)
	}
}

// TODO: find way to test for not touching expired tokens
func testRefreshTokenTouch(t *testing.T, store data.RefreshTokenStore) {
	err := store.Touch(models.RefreshToken("a1b2c3"), 123)
	assert.NoError(t, err)
}

// TODO: find way to test for not finding expired tokens
func testRefreshTokenFindAll(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	// finding nothing
	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Len(t, tokens, 0)

	token, err := store.Create(id)
	require.NoError(t, err)

	// finding something
	tokens2, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens2)
}

func testRefreshTokenCreate(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	token, err := store.Create(id)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens)
}

func testRefreshTokenRevoke(t *testing.T, store data.RefreshTokenStore) {
	id := 123

	err := store.Revoke(models.RefreshToken("a1b2c3"))
	assert.NoError(t, err)

	token, err := store.Create(id)
	require.NoError(t, err)

	found, err := store.Find(token)
	if assert.NoError(t, err) {
		assert.Equal(t, found, id)
	}

	tokens, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Equal(t, []models.RefreshToken{token}, tokens)

	err = store.Revoke(token)
	assert.NoError(t, err)

	found2, err := store.Find(token)
	assert.Empty(t, found2)
	assert.NoError(t, err)

	tokens2, err := store.FindAll(id)
	assert.NoError(t, err)
	assert.Len(t, tokens2, 0)
}
