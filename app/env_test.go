package app_test

import (
	"os"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupURL(t *testing.T) {
	envName := "TEST_LOOKUP_URL"

	t.Run("failing cases", func(t *testing.T) {
		for _, tc := range []struct {
			env string
		}{
			{"host.domain.com"},
			{"host.domain.com/path"},
			{"host.domain.com:1234"},
			{"host.domain.com:1234/path"},
		} {
			err := os.Setenv(envName, tc.env)
			require.NoError(t, err)

			_, err = app.LookupURL(envName)
			assert.Error(t, err, tc.env)
		}
	})

	t.Run("passing cases", func(t *testing.T) {
		for _, tc := range []struct {
			env    string
			scheme string
			port   string
			path   string
		}{
			{"https://host.domain.com", "https", "", ""},
			{"http://host.domain.com:1234/path", "http", "1234", "/path"},
		} {
			err := os.Setenv(envName, tc.env)
			require.NoError(t, err)

			url, err := app.LookupURL(envName)
			require.NoError(t, err)

			require.Equal(t, "host.domain.com", url.Hostname(), tc.env)
			assert.Equal(t, tc.scheme, url.Scheme)
			assert.Equal(t, tc.port, url.Port())
			assert.Equal(t, tc.path, url.Path)
		}
	})

}
