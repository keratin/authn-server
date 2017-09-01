package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/keratin/authn-server/config"
)

type Client struct {
	BaseURL   string
	Modifiers []modder
}

func NewClient(server *httptest.Server) *Client {
	return &Client{server.URL, []modder{}}
}

func (c *Client) Referred(cfg *config.Config) *Client {
	origin := fmt.Sprintf("http://%s", cfg.ApplicationDomains[0].String())
	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.Header.Add("Referer", origin)
			return req
		}),
	}
}

func (c *Client) WithSession(session *http.Cookie) *Client {
	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.AddCookie(session)
			return req
		}),
	}
}

func (c *Client) Authenticated(cfg *config.Config) *Client {
	return &Client{
		c.BaseURL,
		append(c.Modifiers, func(req *http.Request) *http.Request {
			req.SetBasicAuth(cfg.AuthUsername, cfg.AuthPassword)
			return req
		}),
	}
}

func (c *Client) PostForm(path string, form url.Values) (*http.Response, error) {
	body := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.BaseURL, path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) Delete(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.BaseURL, path), nil)
	if err != nil {
		return nil, err
	}
	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.BaseURL, path), nil)
	if err != nil {
		return nil, err
	}
	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) Patch(path string, form url.Values) (*http.Response, error) {
	body := strings.NewReader(form.Encode())
	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s%s", c.BaseURL, path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, mod := range c.Modifiers {
		req = mod(req)
	}

	return http.DefaultClient.Do(req)
}

type modder func(*http.Request) *http.Request
