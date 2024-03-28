package oauth_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestProviderOverrides(t *testing.T) {
	config := &oauth2.Config{ClientSecret: uuid.NewString()}

	t.Run("normal", func(t *testing.T) {
		p := oauth.NewProvider(config, nil)
		secret, err := p.Secret()
		assert.NoError(t, err)
		assert.Equal(t, config.ClientSecret, secret)

		assert.Nil(t, p.AuthCodeOptions())
		assert.Equal(t, http.MethodGet, p.ReturnMethod())
	})

	t.Run("secret override", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			p := oauth.NewProvider(config, nil, oauth.WithSecretGenerator(func() (string, error) {
				return "", assert.AnError
			}))
			secret, err := p.Secret()
			assert.Error(t, err)
			assert.Empty(t, secret)
		})

		t.Run("ok", func(t *testing.T) {
			override := uuid.NewString()
			p := oauth.NewProvider(config, nil, oauth.WithSecretGenerator(func() (string, error) {
				return override, nil
			}))
			secret, err := p.Secret()
			assert.NoError(t, err)
			assert.Equal(t, override, secret)
		})
	})

	t.Run("auth code options", func(t *testing.T) {
		override := []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("x", "y"),
		}
		p := oauth.NewProvider(config, nil, oauth.WithAuthCodeOptions(override...))
		options := p.AuthCodeOptions()
		assert.Equal(t, override, options)
	})

	t.Run("method override", func(t *testing.T) {
		override := uuid.NewString()
		p := oauth.NewProvider(config, nil, oauth.WithReturnMethod(override))
		returnMethod := p.ReturnMethod()
		assert.Equal(t, override, returnMethod)
	})
}
