package data

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/keratin/authn-server/lib"

	"github.com/keratin/authn-server/lib/compat"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// MaintainKeyStore maintains a rotating key store by periodically generating new keys. It will only
// persist keys into an _encrypted_ blob store.
func MaintainKeyStore(ks *RotatingKeyStore, store *EncryptedBlobStore, reporter ops.ErrorReporter, interval time.Duration) error {
	m := &maintainer{
		store: store,
		// the rotation interval should be slightly longer than access token expiry.
		// this means that when a key goes inactive for some interval, we can know
		// that it is useless and discardable by the third interval.
		interval:    interval,
		keyStrength: 2048,
	}
	err := m.maintain(ks, reporter)
	if err != nil {
		return errors.Wrap(err, "maintain")
	}

	return nil
}

type maintainer struct {
	// the rotation interval must be the same length as an access token expiry. that way a key can
	// be in active use for one interval, remain available for verifying old access tokens during
	// the second interval, and be removed and discarded during the third interval.
	interval    time.Duration
	keyStrength int
	store       *EncryptedBlobStore
}

// maintain will restore and rotate a keyStore at periodic intervals.
func (m *maintainer) maintain(ks *RotatingKeyStore, r ops.ErrorReporter) error {
	// fetch current keys
	keys, err := m.restore()
	if err != nil {
		return errors.Wrap(err, "restore")
	}

	// rotate in the previous key
	if keys[0] != nil {
		ks.Rotate(keys[0])

		keyID, _ := compat.KeyID(keys[0].Public())
		log.WithFields(log.Fields{"keyID": keyID}).Info("previous key restored")
	}

	// ensure and rotate in the current key
	if keys[1] != nil {
		ks.Rotate(keys[1])

		keyID, _ := compat.KeyID(keys[1].Public())
		log.WithFields(log.Fields{"keyID": keyID}).Info("current key restored")
	} else {
		newKey, err := m.generate()
		if err != nil {
			return errors.Wrap(err, "generate")
		}
		ks.Rotate(newKey)

		keyID, _ := compat.KeyID(newKey.Public())
		log.WithFields(log.Fields{"keyID": keyID}).Info("current key generated")
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

func (m *maintainer) rotate(ks *RotatingKeyStore) error {
	newKey, err := m.generate()
	if err != nil {
		return errors.Wrap(err, "generate")
	}
	ks.Rotate(newKey)

	keyID, _ := compat.KeyID(newKey.Public())
	log.WithFields(log.Fields{"keyID": keyID}).Info("new key generated")

	return nil
}

// restore will query the blob store for the previous and current keys. It returns keys in the
// proper sorting order, with the newest (current) key in last position. missing keys will leave a
// blank slot, so that the caller may choose what to do.
func (m *maintainer) restore() ([]*rsa.PrivateKey, error) {
	bucket := m.currentBucket()
	keys := make([]*rsa.PrivateKey, 2)

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
func (m *maintainer) generate() (*rsa.PrivateKey, error) {
	bucket := m.currentBucket()
	keyName := fmt.Sprintf("rsa:%d", bucket)
	for {
		// check if another server has created it
		existingKey, err := m.find(bucket)
		if err != nil {
			return nil, err
		}
		if existingKey != nil {
			return existingKey, nil
		}
		// acquire write lock (global mutex)
		success, err := m.store.WLock(keyName)
		if err != nil {
			return nil, err
		}
		if success {
			// create a new key
			key, err := rsa.GenerateKey(rand.Reader, m.keyStrength)
			if err != nil {
				return nil, err
			}
			// store the key
			err = m.store.Write(keyName, keyToBytes(key))
			if err != nil {
				return nil, err
			}
			// return the key
			return key, nil
		}

		// wait and try again
		time.Sleep(50 * time.Millisecond)
	}
}

// find will retrieve and deserialize/decrypt from the blob store
func (m *maintainer) find(bucket int64) (*rsa.PrivateKey, error) {
	blob, err := m.store.Read(fmt.Sprintf("rsa:%d", bucket))
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	if blob == nil {
		return nil, nil
	}

	return bytesToKey(blob), nil
}

func (m *maintainer) currentBucket() int64 {
	return time.Now().Unix() / int64(m.interval/time.Second)
}

func keyToBytes(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func bytesToKey(b []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil
	}
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	return key
}
