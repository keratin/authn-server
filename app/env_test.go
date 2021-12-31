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
			{"vnc://default:mypassword@host.domain.com:6349"},
		} {
			err := os.Setenv(envName, tc.env)
			require.NoError(t, err)

			_, err = app.LookupURL(envName)
			assert.Error(t, err, tc.env)
		}
	})

	t.Run("passing cases", func(t *testing.T) {
		for _, tc := range []struct {
			env      string
			scheme   string
			hostname string
			port     string
			path     string
		}{
			{"http://host.domain.com:1234/path", "http", "host.domain.com", "1234", "/path"},
			{"https://host.domain.com", "https", "host.domain.com", "", ""},
			{"mysql://default:mypassword@host.domain.com:3306", "mysql", "host.domain.com", "3306", ""},
			{"postgres://default:mypassword@host.domain.com:5432", "postgres", "host.domain.com", "5432", ""},
			{"sqlite3:/path/to/file.db", "sqlite3", "", "", "/path/to/file.db"},
			{"redis://default:mypassword@host.domain.com:6349", "redis", "host.domain.com", "6349", ""},
			{"rediss://default:mypassword@host.domain.com:6349", "rediss", "host.domain.com", "6349", ""},
		} {
			err := os.Setenv(envName, tc.env)
			require.NoError(t, err)

			url, err := app.LookupURL(envName)
			require.NoError(t, err)

			require.Equal(t, tc.hostname, url.Hostname(), tc.env)
			assert.Equal(t, tc.scheme, url.Scheme)
			assert.Equal(t, tc.port, url.Port())
			assert.Equal(t, tc.path, url.Path)
		}
	})
}
