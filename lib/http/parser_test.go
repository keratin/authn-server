package http

import (
	"github.com/test-go/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestPayloadHandler(t *testing.T) {
	formValues := make(url.Values)
	formPassword := "some"
	formValues["password"] = []string{formPassword}
	jsonPassword := "another"
	jsonBody := "{\"password\":\"" + jsonPassword + "\"}"

	type payload struct {
		Password string `json:"password"`
	}

	t.Run("Should decode from JSON", func(t *testing.T) {
		request := http.Request{Header: map[string][]string{
			"Content-Type": {"application/json"},
		}}
		request.Body = jsonBodyReader(jsonBody)
		var parsed payload
		err := ParsePayload(&request, &parsed)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, parsed.Password, jsonPassword)
	})

	t.Run("Should decode from form values", func(t *testing.T) {
		request := http.Request{Header: map[string][]string{
			"Content-type": {"application/x-www-form-urlencoded"},
		}}
		request.Form = formValues
		var parsed payload
		err := ParsePayload(&request, &parsed)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, parsed.Password, formPassword)
	})

	t.Run("Should return error for unsupported content type", func(t *testing.T) {
		request := http.Request{Header: map[string][]string{
			"Content-type": {"application/unsupported"},
		}}
		var parsed payload
		err := ParsePayload(&request, &parsed)
		if err == nil {
			t.Error("Should have failed for content type application/unsupported")
			return
		}
		if parseError, ok := err.(ParseError); ok {
			assert.Equal(t, parseError.Code, UnsupportedContentType)
		} else {
			t.Error("expected ParseError instance")
		}

	})

	t.Run("Should try parsing form values when no content type is set in headers ensuring backwards compatibility", func(t *testing.T) {
		request := http.Request{}
		var parsed payload
		err := ParsePayload(&request, &parsed)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, parsed, payload{})
	})

	t.Run("Should not break with weird nil headers", func(t *testing.T) {
		request := http.Request{Header: map[string][]string{
			"Content-Type": nil,
		}}
		var parsed payload
		err := ParsePayload(&request, &parsed)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, parsed, payload{})
	})
}

type body struct {
	content      string
	closeVisited bool
}

func (b body) Read(p []byte) (n int, err error) {
	for key, value := range []byte(b.content) {
		if key >= len(p) {
			return key, nil
		}
		p[key] = value
	}
	return len(b.content), nil
}

func (b *body) Close() error {
	b.closeVisited = true
	return nil
}

func jsonBodyReader(bodyContent string) *body {
	return &body{content: bodyContent}
}
