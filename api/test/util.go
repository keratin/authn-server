package test

import (
	"io/ioutil"
	"net/http"
)

func ReadBody(res *http.Response) []byte {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	return body
}
