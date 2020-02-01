package parse

import (
	"github.com/test-go/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestParsingOnServer will do integration tests of a handler, running on a test Server, that uses Parse method
// to try and parse JSON as well as Form value inputs
func TestParsingOnServer(t *testing.T) {

	t.Run("should return error for unsupported content type", func(t *testing.T) {
		var requestBody struct{ Username string }
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		_, _ = http.Post(ts.URL, "application/unsupported", strings.NewReader(""))

		assert.Equal(t, err.(Error).Code, UnsupportedMediaType)
	})

	t.Run("Should parse Form content", func(t *testing.T) {
		var requestBody struct{ Username string }
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		postForm(ts, url.Values{"username": []string{"aUsername"}})

		assert.Equal(t, requestBody.Username, "aUsername")
		assert.Nil(t, err)
	})

	t.Run("Should not break with unknown Form fields sent", func(t *testing.T) {
		var requestBody struct {
			Username     string
			AnotherField string
		}
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		postForm(ts, url.Values{"username": []string{"aUsername"}, "unknownToServer": []string{"unknown value"}})

		assert.Equal(t, requestBody.Username, "aUsername")
		assert.Nil(t, err)
	})

	t.Run("Should decode from JSON", func(t *testing.T) {
		var requestBody struct{ Username string }
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		postJson(ts, "{\"username\":\"aUsername\"}")

		assert.Equal(t, requestBody.Username, "aUsername")
		assert.Nil(t, err)
	})

	t.Run("Should not break with unknown fields in JSON", func(t *testing.T) {
		var requestBody struct {
			Username    string
			UnusedField string
		}
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		postJson(ts, "{\"username\":\"aUsername\", \"unknown\":\"Unknown to the server\"}")

		assert.Equal(t, requestBody.Username, "aUsername")
		assert.Nil(t, err)
	})

	t.Run("Should return an error for malformed JSON request", func(t *testing.T) {
		var requestBody struct {}
		var err error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = Payload(r, &requestBody)
		}))
		defer ts.Close()
		postJson(ts, "{\"missingVariableContent\":}")

		assert.Equal(t, err.(Error).Code, MalformedInput)
	})

	t.Run("Should not break with nil headers", func(t *testing.T) {
		request := http.Request{Header: map[string][]string{
			"Content-Type": nil,
		}}
		var parsed struct{}
		err := Payload(&request, &parsed)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, parsed, struct{}{})
	})
}

func postJson(ts *httptest.Server, json string) {
	_, _ = http.Post(ts.URL, "application/json", strings.NewReader(json))
}

func postForm(ts *httptest.Server, content url.Values) {
	_, _ = http.Post(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(content.Encode()))
}
