package test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL   string
	Modifiers []Modder
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
