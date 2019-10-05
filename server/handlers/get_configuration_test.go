package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfiguration(t *testing.T) {
	app := &app.App{
		Config: &app.Config{
			AuthNURL: &url.URL{Scheme: "https", Host: "authn.example.com", Path: "/foo"},
		},
		Logger: logrus.New(),
	}
	server := test.Server(app)
	defer server.Close()

	res, err := http.Get(fmt.Sprintf("%s/configuration", server.URL))
	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	data := struct {
		JWKSURI string `json:"jwks_uri"`
	}{}
	json.Unmarshal(body, &data)
	assert.Equal(t, "https://authn.example.com/foo/jwks", data.JWKSURI)
}
