package route

import (
	"bytes"
	"encoding/json"
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
	*http.Client
}

const (
	delete  = "DELETE"
	get     = "GET"
	patch   = "PATCH"
	post    = "POST"
	put     = "PUT"
	options = "OPTIONS"

	contentTypeJSON           = "application/json"
	contentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

// NewClient returns a new Client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client:  http.DefaultClient,
	}
}

// With returns a copy of this client with additional modifiers applied after the existing ones
func (c *Client) With(modifiers ...modder) *Client {
	// Avoid unintended overwriting of the modifiers if backing array is shared between forked Clients
	combinedModifiers := make([]modder, len(c.Modifiers)+len(modifiers))
	n := copy(combinedModifiers, c.Modifiers)
	copy(combinedModifiers[n:], modifiers)
	return &Client{
		BaseURL:   c.BaseURL,
		Modifiers: combinedModifiers,
		Client:    c.Client,
	}
}

// Referred will inject an Origin header into a client's requests.
func (c *Client) Referred(domain *Domain) *Client {
	scheme := "http"
	if domain.Port == "443" {
		scheme = "https"
	}
	origin := fmt.Sprintf("%s://%s", scheme, domain.String())

	return c.With(func(req *http.Request) *http.Request {
		req.Header.Add("Origin", origin)
		return req
	})
}

// WithCookie will inject a Cookie header into a client's requests.
func (c *Client) WithCookie(cookie *http.Cookie) *Client {
	return c.With(func(req *http.Request) *http.Request {
		req.AddCookie(cookie)
		return req
	})
}

// WithClient uses the provided client as the embedded HTTP client
func (c *Client) WithClient(client *http.Client) *Client {
	cpy := c.With()
	c.Client = client
	return cpy
}

// Authenticated will inject HTTP Basic Auth configuration into a client's requests.
func (c *Client) Authenticated(username string, password string) *Client {
	return c.With(func(req *http.Request) *http.Request {
		req.SetBasicAuth(username, password)
		return req
	})
}

// Get issues a GET to the specified path like net/http's Get, but with any modifications
// configured for the current client.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.do(get, contentTypeFormURLEncoded, path, nil)
}

// Delete issues a DELETE to the specified path, with any modifications configured for the current
// client.
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.do(delete, contentTypeFormURLEncoded, path, nil)
}

// DeleteJSON issues a DELETE to the specified path like net/http's Delete, but with any modifications
// configured for the current client and accepting a JSON content.
func (c *Client) DeleteJSON(path string, content map[string]interface{}) (*http.Response, error) {
	marshalled, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return c.do(delete, contentTypeJSON, path, bytes.NewReader(marshalled))
}

// PostForm issues a POST to the specified path like net/http's PostForm, but with any modifications
// configured for the current client.
func (c *Client) PostForm(path string, form url.Values) (*http.Response, error) {
	return c.do(post, contentTypeFormURLEncoded, path, strings.NewReader(form.Encode()))
}

// PostJSON issues a POST to the specified path like net/http's Post, but with any modifications
// configured for the current client and accepting a JSON content.
func (c *Client) PostJSON(path string, content map[string]interface{}) (*http.Response, error) {
	marshalled, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return c.do(post, contentTypeJSON, path, bytes.NewReader(marshalled))
}

// Patch issues a PATCH to the specified path like net/http's PostForm, but with any
// modifications configured for the current client.
func (c *Client) Patch(path string, form url.Values) (*http.Response, error) {
	return c.do(patch, contentTypeFormURLEncoded, path, strings.NewReader(form.Encode()))
}

// PatchJSON issues a PATCH to the specified path like net/http's Post using a JSON content in string format, but with any
// modifications configured for the current client.
func (c *Client) PatchJSON(path string, content string) (*http.Response, error) {
	return c.do(patch, contentTypeJSON, path, strings.NewReader(content))
}

// Preflight issues a CORS OPTIONS request
func (c *Client) Preflight(domain *Domain, verb string, path string) (*http.Response, error) {
	cPreflight := c.Referred(domain).With(func(req *http.Request) *http.Request {
		req.Header.Add("Access-Control-Request-Method", verb)
		return req
	})
	return cPreflight.do(options, contentTypeFormURLEncoded, path, nil)
}

func (c *Client) do(verb string, contentType string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(verb, fmt.Sprintf("%s%s", c.BaseURL, path), body)
	if err != nil {
		return nil, err
	}

	if verb == post || verb == patch || verb == put || verb == delete {
		req.Header.Add("Content-Type", contentType)
	}

	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return c.Do(req)
}
