package route

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type modder func(*http.Request) *http.Request

// Client is a HTTP client similar to net/http, but with a fluent API for modifying requests before
// submission. This can be used to inject headers, cookies, etc.
type Client struct {
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

// NewClient returns a new Client.
func NewClient(baseURL string) *Client {
	return &Client{baseURL, []modder{}}
}

// Referred will inject an Origin header into a client's requests.
func (c *Client) Referred(domain *Domain) *Client {
	scheme := "http"
	if domain.Port == "443" {
		scheme = "https"
	}
	origin := fmt.Sprintf("%s://%s", scheme, domain.String())

	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.Header.Add("Origin", origin)
			return req
		}),
	}
}

// WithCookie will inject a Cookie header into a client's requests.
func (c *Client) WithCookie(cookie *http.Cookie) *Client {
	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.AddCookie(cookie)
			return req
		}),
	}
}

// Authenticated will inject HTTP Basic Auth configuration into a client's requests.
func (c *Client) Authenticated(username string, password string) *Client {
	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.SetBasicAuth(username, password)
			return req
		}),
	}
}

// Get issues a GET to the specified path like net/http's Get, but with any modifications
// configured for the current client.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.do(get, path, nil)
}

// Delete issues a DELETE to the specified path, with any modifications configured for the current
// client.
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.do(delete, path, nil)
}

// PostForm issues a POST to the specified path like net/http's PostForm, but with any modifications
// configured for the current client.
func (c *Client) PostForm(path string, form url.Values) (*http.Response, error) {
	return c.do(post, path, strings.NewReader(form.Encode()))
}

// Patch issues a PATCH to the specified path like net/http's PostForm, but with any
// modifications configured for the current client.
func (c *Client) Patch(path string, form url.Values) (*http.Response, error) {
	return c.do(patch, path, strings.NewReader(form.Encode()))
}

func (c *Client) do(verb string, path string, body io.Reader) (*http.Response, error) {
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
