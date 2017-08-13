package compat_test

import (
	"testing"

	"github.com/keratin/authn-server/compat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalString(t *testing.T) {
	testCases := []struct {
		str string
		byt []byte
	}{
		{"hello", []byte("\x04\x08I\"\x0ahello\x06:\x06ET")},
		{"lorem ipsum sit dolor", []byte("\x04\x08I\"\x1alorem ipsum sit dolor\x06:\x06ET")},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.byt, compat.MarshalString(tc.str))

		r, e := compat.UnmarshalString(tc.byt)
		require.NoError(t, e)
		assert.Equal(t, tc.str, string(r))
	}
}
