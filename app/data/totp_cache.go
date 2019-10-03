package data

import (
	"fmt"

	"github.com/pkg/errors"
)

type TOTPCache struct {
	ebs *EncryptedBlobStore
}

func NewTOTPCache(ebs *EncryptedBlobStore) TOTPCache {
	return TOTPCache{
		ebs: ebs,
	}
}

func (t *TOTPCache) CacheTOTPSecret(accountID int, secret []byte) error {
	keyName := fmt.Sprintf("totp:%d", accountID)
	_, err := t.ebs.Write(keyName, secret)
	if err != nil {
		return errors.Wrap(err, "CacheTOTPSecret")
	}
	return nil
}

func (t *TOTPCache) LoadTOTPSecret(accountID int) ([]byte, error) {
	keyName := fmt.Sprintf("totp:%d", accountID)
	val, err := t.ebs.Read(keyName)
	if err != nil {
		return nil, errors.Wrap(err, "LoadTOTPSecret")
	}
	return val, nil
}
