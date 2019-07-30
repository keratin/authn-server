package handlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealth(t *testing.T) {
	app := &app.App{
		DbCheck:    func() bool { return true },
		RedisCheck: func() bool { return true },
		Config:     &app.Config{},
		Logger:     logrus.New(),
	}
	server := test.Server(app)
	defer server.Close()

	res, err := http.Get(fmt.Sprintf("%s/health", server.URL))
	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	assert.Equal(t, `{"http":true,"db":true,"redis":true}`, string(body))
}
