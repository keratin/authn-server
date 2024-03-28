package apple

import (
	"fmt"
	"strconv"
)

func ExtractCredentialData(credData map[string]string) (string, string, int64, error) {
	var (
		keyID        string
		teamID       string
		expiresInStr string
		expiresIn    int64
		found        bool
		constructErr error
	)

	if keyID, found = credData["keyID"]; !found || keyID == "" {
		return "", "", 0, fmt.Errorf("missing keyID")
	}

	if teamID, found = credData["teamID"]; !found || teamID == "" {
		return "", "", 0, fmt.Errorf("missing teamID")
	}

	if expiresInStr, found = credData["expirySeconds"]; !found || expiresInStr == "" {
		return "", "", 0, fmt.Errorf("missing expirySeconds")
	} else {
		expiresIn, constructErr = strconv.ParseInt(expiresInStr, 10, 0)
		if constructErr != nil {
			return "", "", 0, fmt.Errorf("failed to parse expirySeconds: %w", constructErr)
		}
	}

	return teamID, keyID, expiresIn, nil
}
