package route_test

import (
	"net/url"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomain(t *testing.T) {
	h := "host"
	hp := "host:port"
	hDomain := &route.Domain{Hostname: "host"}
	hpDomain := &route.Domain{Hostname: "host", Port: "port"}

	t.Run("ParseDomain", func(t *testing.T) {
		assert.Equal(t, "host", route.ParseDomain(h).Hostname)
		assert.Equal(t, "", route.ParseDomain(h).Port)

		assert.Equal(t, "host", route.ParseDomain(hp).Hostname)
		assert.Equal(t, "port", route.ParseDomain(hp).Port)
	})

	t.Run("Matches", func(t *testing.T) {
		testCases := []struct {
			domain  string
			url     string
			matches bool
		}{
			{"example.com", "http://example.com", true},
			{"example.com", "http://example.com:3000", true},
			{"www.example.com", "http://www.example.com", true},
			{"www.example.com", "http://example.com", false},
			{"example.com:3000", "http://example.com:3000", true},
			{"example.com:3000", "http://example.com:8080", false},
			{"example.com:80", "http://example.com", true},
			{"example.com:80", "https://example.com", false},
			{"example.com:80", "http://example.com:3000", false},
			{"example.com:443", "https://example.com", true},
			{"example.com:443", "http://example.com", false},
			{"example.com:443", "https://example.com:3000", false},
		}

		for _, tc := range testCases {
			d := route.ParseDomain(tc.domain)
			u, err := url.Parse(tc.url)
			require.NoError(t, err)

			assert.Equal(t, tc.matches, d.Matches(u))
		}
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "host", hDomain.String())
		assert.Equal(t, "host:port", hpDomain.String())
	})
}
