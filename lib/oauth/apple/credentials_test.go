package apple_test

import (
	"testing"

	"github.com/keratin/authn-server/lib/oauth/apple"
	"github.com/stretchr/testify/assert"
)

func TestExtractCredentialData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		teamID, keyID, expiresIn, err := apple.ExtractCredentialData(map[string]string{
			"keyID":         "keyID",
			"teamID":        "teamID",
			"expirySeconds": "3600",
		})
		assert.NoError(t, err)
		assert.Equal(t, "teamID", teamID)
		assert.Equal(t, "keyID", keyID)
		assert.Equal(t, int64(3600), expiresIn)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("keyID", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{})
				assert.EqualError(t, err, "missing keyID")
			})
			t.Run("empty", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": ""})
				assert.EqualError(t, err, "missing keyID")
			})
		})

		t.Run("teamID", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": "keyID"})
				assert.EqualError(t, err, "missing teamID")
			})
			t.Run("empty", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": "keyID", "teamID": ""})
				assert.EqualError(t, err, "missing teamID")
			})
		})

		t.Run("expirySeconds", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": "keyID", "teamID": "teamID"})
				assert.EqualError(t, err, "missing expirySeconds")
			})
			t.Run("empty", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": "keyID", "teamID": "teamID", "expirySeconds": ""})
				assert.EqualError(t, err, "missing expirySeconds")
			})
			t.Run("invalid", func(t *testing.T) {
				_, _, _, err := apple.ExtractCredentialData(map[string]string{"keyID": "keyID", "teamID": "teamID", "expirySeconds": "invalid"})
				assert.EqualError(t, err, "failed to parse expirySeconds: strconv.ParseInt: parsing \"invalid\": invalid syntax")
			})
		})
	})
}
