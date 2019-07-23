package data

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/keratin/authn-server/app/data/private"
	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewKeyStoreRotater creates a KeyStoreRotater.
//
// The rotation interval should match the lifetime of an access token. This means a key can be used
// to sign tokens for one time period, remain available to verify tokens for another time period,
// and be discarded during the third.
func NewKeyStoreRotater(blobStore *EncryptedBlobStore, interval time.Duration, logger logrus.FieldLogger) *KeyStoreRotater {
	return &KeyStoreRotater{
		store:       blobStore,
		interval:    interval,
		keyStrength: 2048,
		logger:      logger.WithField("scope", "NewKeyStoreRotater"),
	}
}

// KeyStoreRotater will rotate a RotatingKeyStore by periodically generating new keys. The keys will be
// persisted into an EncryptedBlobStore, shared with other processes, and read back on startup.
type KeyStoreRotater struct {
	interval    time.Duration
	keyStrength int
	store       *EncryptedBlobStore
	logger      logrus.FieldLogger
}

// Maintain will restore and rotate a keyStore at periodic intervals. It will return an error only
// for issues during startup. Any issues that arise later during background work will be reported.
func (m *KeyStoreRotater) Maintain(ks *RotatingKeyStore, r ops.ErrorReporter) error {
	// fetch current keys
	keys, err := m.restore()
	if err != nil {
		return errors.Wrap(err, "restore")
	}

	// rotate in the previous key
	if keys[0] != nil {
		ks.Rotate(keys[0])

		m.logger.WithField("keyID", keys[0].JWK.KeyID).Info("previous key restored")
	}

	// ensure and rotate in the current key
	if keys[1] != nil {
		ks.Rotate(keys[1])

		m.logger.WithField("keyID", keys[1].JWK.KeyID).Info("current key restored")
	} else {
		newKey, err := m.generate()
		if err != nil {
			return errors.Wrap(err, "generate")
		}
		ks.Rotate(newKey)
	}

	go func() {
		intervals := lib.EpochIntervalTick(m.interval)
		for range intervals {
			err = m.rotate(ks)
			if err != nil {
				r.ReportError(err)
			}
		}
	}()

	return nil
}

func (m *KeyStoreRotater) rotate(ks *RotatingKeyStore) error {
	newKey, err := m.generate()
	if err != nil {
		return errors.Wrap(err, "generate")
	}
	ks.Rotate(newKey)

	return nil
}

// restore will query the blob store for the previous and current keys. It returns keys in the
// proper sorting order, with the newest (current) key in last position. missing keys will leave a
// blank slot, so that the caller may choose what to do.
func (m *KeyStoreRotater) restore() ([]*private.Key, error) {
	bucket := m.currentBucket()
	keys := make([]*private.Key, 2)

	previous, err := m.find(bucket - 1)
	if err != nil {
		return nil, err
	}
	keys[0] = previous

	current, err := m.find(bucket)
	if err != nil {
		return nil, err
	}
	keys[1] = current

	return keys, nil
}

// generate will create a new key and store it as an encrypted blob. It relies on a write lock to
// coordinate with other AuthN servers.
func (m *KeyStoreRotater) generate() (*private.Key, error) {
	keyName := fmt.Sprintf("rsa:%d", m.currentBucket())
	key, err := private.GenerateKey(m.keyStrength)
	if err != nil {
		return nil, err
	}

	blob := keyToBytes(key)
	ok, err := m.store.WriteNX(keyName, blob)
	if err != nil {
		return nil, err
	}

	if ok {
		m.logger.WithFields(logrus.Fields{"keyID": key.JWK.KeyID, "keyName": keyName}).Info("new key generated")
	} else {
		keyBlob, err := m.store.Read(keyName)
		if err != nil {
			return nil, err
		}
		key = bytesToKey(keyBlob)

		m.logger.WithField("keyID", key.JWK.KeyID).Info("key synchronized")
	}

	return key, nil
}

// find will retrieve and deserialize/decrypt from the blob store
func (m *KeyStoreRotater) find(bucket int64) (*private.Key, error) {
	blob, err := m.store.Read(fmt.Sprintf("rsa:%d", bucket))
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	if blob == nil {
		return nil, nil
	}

	return bytesToKey(blob), nil
}

func (m *KeyStoreRotater) currentBucket() int64 {
	return time.Now().Unix() / int64(m.interval/time.Second)
}

func keyToBytes(key *private.Key) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key.PrivateKey),
	})
}

func bytesToKey(b []byte) *private.Key {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil
	}
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	privateKey, _ := private.NewKey(key)
	return privateKey
}
