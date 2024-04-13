package services_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/oauth"
	"golang.org/x/oauth2"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
)

func TestIdentityReconciler(t *testing.T) {
	store := mock.NewAccountStore()
	cfg := &app.Config{}

	t.Run("linked account", func(t *testing.T) {
		acct, err := store.Create("linked@test.com", []byte("password"))
		require.NoError(t, err)
		err = store.AddOauthAccount(acct.ID, "testProvider", "123", "email", "TOKEN")
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "123", Email: "linked@test.com"}, &oauth2.Token{}, 0)
		assert.NoError(t, err)
		if assert.NotNil(t, found) {
			assert.Equal(t, found.Username, "linked@test.com")
		}
	})

	t.Run("linked account that is locked", func(t *testing.T) {
		acct, err := store.Create("linkedlocked@test.com", []byte("password"))
		require.NoError(t, err)
		err = store.AddOauthAccount(acct.ID, "testProvider", "234", "email", "TOKEN")
		require.NoError(t, err)
		_, err = store.Lock(acct.ID)
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "234", Email: "linkedlocked@test.com"}, &oauth2.Token{}, 0)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("linkable account", func(t *testing.T) {
		acct, err := store.Create("linkable@test.com", []byte("password"))
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "345", Email: "linkable@test.com"}, &oauth2.Token{}, acct.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, found) {
			assert.Equal(t, found.Username, "linkable@test.com")
		}
	})

	t.Run("linkable account that is linked", func(t *testing.T) {
		acct, err := store.Create("linkablelinked@test.com", []byte("password"))
		require.NoError(t, err)
		err = store.AddOauthAccount(acct.ID, "testProvider", "0", "email", "TOKEN")
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "456", Email: "linkablelinked@test.com"}, &oauth2.Token{}, acct.ID)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("new account", func(t *testing.T) {
		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "567", Email: "new@test.com"}, &oauth2.Token{}, 0)
		assert.NoError(t, err)
		if assert.NotNil(t, found) {
			assert.Equal(t, found.Username, "new@test.com")
		}
	})

	t.Run("new account with username collision", func(t *testing.T) {
		_, err := store.Create("existing@test.com", []byte("password"))
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, "testProvider", &oauth.UserInfo{ID: "678", Email: "existing@test.com"}, &oauth2.Token{}, 0)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("update missing email after oauth migration table", func(t *testing.T) {
		provider := "testProvider"
		providerAccountId := "666"
		email := "update-missing-oauth-email@test.com"

		account, err := store.Create(email, []byte("password"))
		require.NoError(t, err)

		err = store.AddOauthAccount(account.ID, provider, providerAccountId, "", "TOKEN")
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, provider, &oauth.UserInfo{ID: providerAccountId, Email: email}, &oauth2.Token{}, 0)
		assert.NoError(t, err)
		assert.NotNil(t, found)

		oAccounts, err := store.GetOauthAccounts(account.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(oAccounts))
		assert.Equal(t, email, oAccounts[0].Email)
	})

	t.Run("update oauth email when is outdated", func(t *testing.T) {
		provider := "testProvider"
		providerAccountId := "777"
		email := "update-outdate-oauth-email@test.com"

		account, err := store.Create(email, []byte("password"))
		require.NoError(t, err)

		err = store.AddOauthAccount(account.ID, provider, providerAccountId, "email@email.com", "TOKEN")
		require.NoError(t, err)

		found, err := services.IdentityReconciler(store, cfg, provider, &oauth.UserInfo{ID: providerAccountId, Email: email}, &oauth2.Token{}, 0)
		assert.NoError(t, err)
		assert.NotNil(t, found)

		oAccounts, err := store.GetOauthAccounts(account.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(oAccounts))
		assert.Equal(t, email, oAccounts[0].Email)
	})
}
