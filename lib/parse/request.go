package parse

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"io"
	"net/http"
	"strings"
)

const (
	applicationJson           = "application/json"
	applicationFormUrlEncoded = "application/x-www-form-urlencoded"

	UnsupportedMediaType = ErrorCode(1)
	MalformedInput       = ErrorCode(2)
)

type ErrorCode int

type Error struct {
	Message string
	Code    ErrorCode
}

func (e Error) String() string {
	return fmt.Sprintf("Payload parse error: %s", e.Message)
}

func (e Error) Error() string {
	return e.String()
}

// Payload parses a request body, depending on the type set in Content-Type found in the headers.
// If no Content-Type is set in the headers, Payload will try to parse content from sent Form values.
func Payload(r *http.Request, value interface{}) error {
	contentType := strings.ToLower(getContentType(r.Header))
	if contentType == "" {
		contentType = applicationFormUrlEncoded
	}

	switch {
	case strings.Contains(contentType, applicationJson):
		return parseJson(r.Body, value)
	case strings.Contains(contentType, applicationFormUrlEncoded):
		return parseForm(value, r)
	}

	return Error{Code: UnsupportedMediaType, Message: fmt.Sprintf("Unsupported Content-Type '%s'", contentType)}
}

func parseForm(value interface{}, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return Error{
			Message: err.Error(),
			Code:    MalformedInput,
		}
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	return decoder.Decode(value, r.Form)
}

func parseJson(r io.ReadCloser, parsed interface{}) error {
	err := json.NewDecoder(r).Decode(parsed)
	defer r.Close()

	return err
}

func getContentType(headers http.Header) string {
	for key, value := range headers {
		if strings.Contains(strings.ToLower(key), "content-type") && len(value) > 0 {
			return value[0]
		}
	}
	return ""
}
