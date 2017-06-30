package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

type reqModder func(req *http.Request) *http.Request

func Get(path string, h http.HandlerFunc, befores ...reqModder) *httptest.ResponseRecorder {
	return makeRequest("GET", path, h, nil, befores...)
}

func Delete(path string, h http.HandlerFunc, befores ...reqModder) *httptest.ResponseRecorder {
	return makeRequest("DELETE", path, h, nil, befores...)
}

func Post(path string, h http.HandlerFunc, params map[string]string, befores ...reqModder) *httptest.ResponseRecorder {
	return makeRequest("POST", path, h, strings.NewReader(mapToParams(params)), befores...)
}

func makeRequest(verb string, path string, h http.HandlerFunc, body io.Reader, befores ...reqModder) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(verb, "/health", body)
	if body != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, before := range befores {
		req = before(req)
	}

	h.ServeHTTP(res, req)
	return res
}

func mapToParams(params map[string]string) string {
	buffer := make([]string, 0)
	for k, v := range params {
		buffer = append(buffer, strings.Join([]string{k, v}, "="))
	}
	return strings.Join(buffer, "&")
}
