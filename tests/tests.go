package tests

import (
	"reflect"
	"testing"
)

func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("\nexpected: %T %v\n     got: %T %v", expected, expected, actual, actual)
	}
}

func RefuteEqual(t *testing.T, unexpected interface{}, actual interface{}) {
	if reflect.DeepEqual(unexpected, actual) {
		t.Errorf("\nunexpected value: %T %v", unexpected, unexpected)
	}
}
