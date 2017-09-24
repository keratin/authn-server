package route

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type modder func(*http.Request) *http.Request

type client struct {
	BaseURL   string
	Modifiers []modder
}

const (
	delete = "DELETE"
	get    = "GET"
	patch  = "PATCH"
	post   = "POST"
	put    = "PUT"
)

// NewClient returns a HTTP client similar to net/http's but with a fluent API for modifying the
// request with headers before submission.
func NewClient(baseURL string) *client {
	return &client{baseURL, []modder{}}
}

// Referred will inject a Referer header into a client's requests.
func (c *client) Referred(domain *Domain) *client {
	scheme := "http"
	if domain.Port == "443" {
		scheme = "https"
	}
	origin := fmt.Sprintf("%s://%s", scheme, domain.String())

	return &client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.Header.Add("Referer", origin)
			return req
		}),
	}
}

// WithCookie will inject a Cookie header into a client's requests.
func (c *client) WithCookie(cookie *http.Cookie) *client {
	return &client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.AddCookie(cookie)
			return req
		}),
	}
}

// Authenticated will inject HTTP Basic Auth configuration into a client's requests.
func (c *client) Authenticated(username string, password string) *client {
	return &client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.SetBasicAuth(username, password)
			return req
		}),
	}
}

// PostForm issues a GET to the specified path like net/http's Get, but with any modifications
// configured for the current client.
func (c *client) Get(path string) (*http.Response, error) {
	return c.do(get, path, nil)
}

// PostForm issues a DELETE to the specified path, with any modifications configured for the current
// client.
func (c *client) Delete(path string) (*http.Response, error) {
	return c.do(delete, path, nil)
}

// PostForm issues a POST to the specified path like net/http's PostForm, but with any modifications
// configured for the current client.
func (c *client) PostForm(path string, form url.Values) (*http.Response, error) {
	return c.do(post, path, strings.NewReader(form.Encode()))
}

// PostForm issues a PATCH to the specified path like net/http's PostForm, but with any
// modifications configured for the current client.
func (c *client) Patch(path string, form url.Values) (*http.Response, error) {
	return c.do(patch, path, strings.NewReader(form.Encode()))
}

func (c *client) do(verb string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(verb, fmt.Sprintf("%s%s", c.BaseURL, path), body)

	if verb == post || verb == patch || verb == put {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, err
	}
	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return http.DefaultClient.Do(req)
}
