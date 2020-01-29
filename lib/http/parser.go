package http

import (
	"fmt"
	"github.com/gorilla/schema"
	"gopkg.in/square/go-jose.v2/json"
	"io"
	"net/http"
	"strings"
)

const (
	applicationJson           = "application/json"
	applicationFormUrlEncoded = "application/x-www-form-urlencoded"

	UnsupportedContentType = ParseErrorCode(1)
)

type ParseErrorCode int

type ParseError struct {
	Message string
	Code    ParseErrorCode
}

func (e ParseError) String() string {
	return fmt.Sprintf("Payload parse error: %s", e.Message)
}

func (e ParseError) Error() string {
	return e.String()
}

// ParsePayload parses a request body, depending on the type set in Content-Type found in the headers.
// If no Content-Type is set in the headers, ParsePayload will try to parse content from sent Form values.
func ParsePayload(r *http.Request, value interface{}) error {
	contentType := strings.ToLower(getContentType(r.Header))
	if contentType == "" {
		contentType = applicationFormUrlEncoded
	}

	switch {
	case strings.Contains(contentType, applicationJson):
		return parseJson(r.Body, value)
	case strings.Contains(contentType, applicationFormUrlEncoded):
		return schema.NewDecoder().Decode(value, r.Form)
	}

	return ParseError{Code: UnsupportedContentType, Message: fmt.Sprintf("Unsupported Content-Type '%s'", contentType)}
}

func parseJson(r io.ReadCloser, parsed interface{}) error {
	err := json.NewDecoder(r).Decode(parsed)
	defer r.Close()

	return err
}

func getContentType(headers http.Header) string {
	for key, value := range headers	{
		if strings.Contains(strings.ToLower(key), "content-type") && len(value) > 0 {
			return value[0]
		}
	}
	return ""
}

