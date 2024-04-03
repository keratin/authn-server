package apple

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

type appleKeyResponse struct {
	Keys []appleKey `json:"keys"`
}

type appleKey struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

type KeyNotFoundError struct {
	KeyID string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key %s not found", e.KeyID)
}

type rsaKeyStore interface {
	Get(keyID string) (*rsa.PublicKey, error)
}

func newSigningKeyStore(client *http.Client) *signingKeyStore {
	return &signingKeyStore{
		client: client,
		keys:   make(map[string]*rsa.PublicKey),
	}
}

// signingKeyStore provides a store for apple public keys.
type signingKeyStore struct {
	client *http.Client
	keys   map[string]*rsa.PublicKey
}

func (a *signingKeyStore) Get(keyID string) (*rsa.PublicKey, error) {
	itm, got := a.keys[keyID]
	if got {
		return itm, nil
	}

	err := a.refresh()

	if err != nil {
		return nil, fmt.Errorf("failed to refresh apple keys: %w", err)
	}

	itm, got = a.keys[keyID]
	if got {
		return itm, nil
	}

	return nil, &KeyNotFoundError{KeyID: keyID}
}

func (a *signingKeyStore) refresh() error {
	var x appleKeyResponse
	_ = x.Keys
	keysResp, keysErr := a.client.Get("https://appleid.apple.com/auth/keys")
	if keysErr != nil {
		return fmt.Errorf("failed to fetch apple keys: %w", keysErr)
	}

	keysBody, keysErr := io.ReadAll(keysResp.Body)
	if keysErr != nil {
		return fmt.Errorf("failed to read apple keys: %w", keysErr)
	}

	keys := appleKeyResponse{}
	keysErr = json.Unmarshal(keysBody, &keys)

	if keysErr != nil {
		return fmt.Errorf("failed to unmarshal apple keys: %w", keysErr)
	}

	newKeys := make(map[string]*rsa.PublicKey, len(keys.Keys))

	for _, key := range keys.Keys {
		var err error
		// build key and place in new map
		publicKey := new(rsa.PublicKey)
		publicKey.N, err = decodeBase64BigInt(key.N)
		if err != nil {
			return fmt.Errorf("failed to decode N for key %s: %w", key.Kid, err)
		}

		var e *big.Int
		e, err = decodeBase64BigInt(key.E)
		if err != nil {
			return fmt.Errorf("failed to decode E for key %s: %w", key.Kid, err)
		}

		publicKey.E = int(e.Int64())

		newKeys[key.Kid] = publicKey
	}

	a.keys = newKeys

	return nil
}

func decodeBase64BigInt(s string) (*big.Int, error) {
	buffer, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	return big.NewInt(0).SetBytes(buffer), nil
}
