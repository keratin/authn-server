package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keratin/authn-server/api"
)

func ReadBody(res *http.Response) []byte {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	return body
}

// ExtractResult reads the value from inside a successful result envelope. It must be provided with
// `inner`, an empty struct that describes the expected (desired) shape of what is inside the
// envelope.
func ExtractResult(res *http.Response, inner interface{}) error {
	return json.Unmarshal(ReadBody(res), &api.ServiceData{inner})
}
