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
	origin := fmt.Sprintf("http://%s", cfg.ApplicationDomains[0])
	c.Modifiers = append(c.Modifiers, func(req *http.Request) *http.Request {
		req.Header.Add("Referer", origin)
		return req
	})
	return c
}

func (c *Client) WithSession(session *http.Cookie) *Client {
	c.Modifiers = append(c.Modifiers, func(req *http.Request) *http.Request {
		req.AddCookie(session)
		return req
	})
	return c
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

type modder func(req *http.Request) *http.Request
