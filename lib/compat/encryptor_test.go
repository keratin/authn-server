package compat_test

import (
	"encoding/hex"
	"testing"

	"github.com/keratin/authn-server/lib/compat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAndDecrypt(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	nonce, _ := hex.DecodeString("37b8e8a308c354048d245f6d")

	testCases := []struct {
		plaintext string
		message   string
	}{
		{"exampleplaintext", "cWmCKah1H16f7AUzjZS3BAzglkH3Wc/BOWM=--N7joowjDVASNJF9t--OXIx+Gcse9c+/zCfUusrMQ=="},
	}

	for _, tc := range testCases {
		msg, err := compat.EncryptWithNonce([]byte(tc.plaintext), key, nonce)
		require.NoError(t, err)
		assert.Equal(t, tc.message, string(msg))

		data, err := compat.Decrypt(msg, key)
		require.NoError(t, err)
		assert.Equal(t, tc.plaintext, string(data))
	}
}
