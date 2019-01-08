package cors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keratin/authn-server/server/cors"
	"github.com/keratin/authn-server/lib/route"
)

func TestOriginValidator(t *testing.T) {
	validator := cors.OriginValidator([]route.Domain{
		route.ParseDomain("example.com:443"),
	})

	assert.True(t, validator("https://example.com"))
	assert.False(t, validator("http://example.com"))
	assert.False(t, validator("https://example.com:9100"))
	assert.False(t, validator("https://www.example.com"))
}
