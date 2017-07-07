package test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keratin/authn-server/config"
)

func ReferFrom(cfg *config.Config) Modder {
	origin := fmt.Sprintf("http://%s", cfg.ApplicationDomains[0])
	return func(req *http.Request) *http.Request {
		req.Header.Add("Referer", origin)
		return req
	}
}

func WithSession(session *http.Cookie) Modder {
	return func(req *http.Request) *http.Request {
		req.AddCookie(session)
		return req
	}
}

type Modder func(req *http.Request) *http.Request

func ReadBody(res *http.Response) []byte {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	return body
}
