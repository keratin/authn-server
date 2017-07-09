package services

import (
	"fmt"
	"strings"
)

var ErrMissing = "MISSING"
var ErrTaken = "TAKEN"
var ErrFormatInvalid = "FORMAT_INVALID"
var ErrInsecure = "INSECURE"
var ErrFailed = "FAILED"
var ErrLocked = "LOCKED"
var ErrExpired = "EXPIRED"
var ErrNotFound = "NOT_FOUND"

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e fieldError) String() string {
	return fmt.Sprintf("%v: %v", e.Field, e.Message)
}

type FieldErrors []fieldError

func (es FieldErrors) Error() string {
	var buf = make([]string, len(es))
	for _, e := range es {
		buf = append(buf, e.String())
	}
	return strings.Join(buf, ", ")
}
