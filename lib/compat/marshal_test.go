package compat_test

import (
	"strings"
	"testing"

	"github.com/keratin/authn-server/lib/compat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalInt(t *testing.T) {
	testCases := []struct {
		input  int
		output []byte
	}{
		{0, []byte{0x04, 0x08, 0x69, 0x05}},
		{1, []byte{0x04, 0x08, 0x69, 0x06}},
		{122, []byte{0x04, 0x08, 0x69, 0x7f}},
		{123, []byte{0x04, 0x08, 0x69, 0x01, 0x7b}},
		{255, []byte{0x04, 0x08, 0x69, 0x01, 0xff}},
		{256, []byte{0x04, 0x08, 0x69, 0x02, 0x00, 0x01}},
		{65535, []byte{0x04, 0x08, 0x69, 0x02, 0xff, 0xff}},
		{65536, []byte{0x04, 0x08, 0x69, 0x03, 0x00, 0x00, 0x01}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.output, compat.Marshal(tc.input))

		i, e := compat.UnmarshalInt(tc.output)
		require.NoError(t, e)
		assert.Equal(t, tc.input, i)
	}
}

func TestMarshalString(t *testing.T) {
	shortStr := "lorem ipsum sit dolor"
	longStr := strings.Repeat("1234567890", 15)

	testCases := []struct {
		str string
		byt []byte
	}{
		{shortStr, []byte("\x04\x08I\"\x1a" + shortStr + "\x06:\x06ET")},
		{longStr, []byte("\x04\x08I\"\x01\x96" + longStr + "\x06:\x06ET")},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.byt, compat.Marshal(tc.str))

		r, e := compat.UnmarshalString(tc.byt)
		require.NoError(t, e)
		assert.Equal(t, tc.str, string(r))
	}
}
