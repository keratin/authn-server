package compat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// Encrypt performs AES-256-GCM encryption in a Rails-compatible manner by imitating the
// `Marshal.dump(value)`` effect (for a string value, as of v4.8) and also splitting the
// authentication tag into a separate value joined with dashes.
//
// See: ActiveSupport::MessageEncryptor with GCM changes
func Encrypt(value []byte, secret []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "ReadFull")
	}
	return EncryptWithNonce(value, secret, nonce)
}

// EncryptWithNonce is the deterministic core of Encrypt for testing purposes.
func EncryptWithNonce(value []byte, secret []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, errors.Wrap(err, "NewCipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "NewGCM")
	}

	ciphertext := aesgcm.Seal(nil, nonce, Marshal(string(value)), nil)
	encryptedData := ciphertext[:len(ciphertext)-16]
	authTag := ciphertext[len(ciphertext)-16:]

	return []byte(fmt.Sprintf(
		"%v--%v--%v",
		base64.StdEncoding.EncodeToString(encryptedData),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(authTag),
	)), nil
}

// Decrypt performs AES-256-GCM decryption in a Rails-compatible manner by expecting the message to
// comprise encrypted data, nonce, and an auth tag. It will also attempt to unwrap the plaintext as
// if it were Marshal.dump'd.
//
// See: ActiveSupport::MessageEncryptor with GCM changes
func Decrypt(message []byte, key []byte) (string, error) {
	slices := strings.Split(string(message), "--")
	encryptedData, _ := base64.StdEncoding.DecodeString(slices[0])
	nonce, _ := base64.StdEncoding.DecodeString(slices[1])
	authTag, _ := base64.StdEncoding.DecodeString(slices[2])
	if authTag == nil || len(authTag) != 16 {
		return "", fmt.Errorf("unexpected encrypted message format")
	}

	ciphertext := bytes.Join([][]byte{encryptedData, authTag}, []byte(""))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "NewCipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "NewGCM")
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.Wrap(err, "Open")
	}

	return UnmarshalString(plaintext)
}
