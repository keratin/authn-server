package data

import (
	"fmt"

	"github.com/pkg/errors"
)

type TOTPCache interface {
	CacheTOTPSecret(accountID int, secret []byte) error
	LoadTOTPSecret(accountID int) ([]byte, error)
}
type totpCache struct {
	ebs *EncryptedBlobStore
}

func NewTOTPCache(ebs *EncryptedBlobStore) TOTPCache {
	return &totpCache{
		ebs: ebs,
	}
}

func (t *totpCache) CacheTOTPSecret(accountID int, secret []byte) error {
	keyName := fmt.Sprintf("totp:%d", accountID)
	_, err := t.ebs.Write(keyName, secret)
	if err != nil {
		return errors.Wrap(err, "CacheTOTPSecret")
	}
	return nil
}

func (t *totpCache) LoadTOTPSecret(accountID int) ([]byte, error) {
	keyName := fmt.Sprintf("totp:%d", accountID)
	val, err := t.ebs.Read(keyName)
	if err != nil {
		return nil, errors.Wrap(err, "LoadTOTPSecret")
	}
	return val, nil
}
