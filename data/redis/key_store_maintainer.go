package redis

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/compat"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
)

type maintainer struct {
	// the rotation interval should be slightly longer than access token expiry.
	// this means that when a key goes inactive for some interval, we can know
	// that it is useless and discardable by the third interval.
	interval time.Duration

	// if two clients need to regenerate a key at the same time, this is how long
	// one will have to attempt it while the other waits patiently.
	//
	// this should be greater than the peak time necessary to generate and encrypt a
	// key, plus send it back over the wire to redis.
	race time.Duration

	keyStrength   int
	encryptionKey []byte
	client        *redis.Client
}

// maintain will restore and rotate a keyStore at periodic intervals.
func (m *maintainer) maintain(ks *keyStore, r ops.ErrorReporter) error {
	// fetch current keys
	keys, err := m.restore()
	if err != nil {
		return errors.Wrap(err, "restore")
	}

	// rotate in the previous key
	if keys[0] != nil {
		ks.Rotate(keys[0])
	}

	// ensure and rotate in the current key
	if keys[1] != nil {
		ks.Rotate(keys[1])
	} else {
		newKey, err := m.generate()
		if err != nil {
			return errors.Wrap(err, "generate")
		}
		ks.Rotate(newKey)
	}

	go func() {
		ticker := NewEpochIntervalTicker(m.interval)
		for range ticker {
			err = m.rotate(ks)
			if err != nil {
				r.ReportError(err)
			}
		}
	}()

	return nil
}

func (m *maintainer) rotate(ks *keyStore) error {
	newKey, err := m.generate()
	if err != nil {
		return errors.Wrap(err, "generate")
	}
	ks.Rotate(newKey)
	return nil
}

// restore will query Redis for the previous and current keys. It returns keys in the proper sorting
// order, with the newest (current) key in last position. missing keys will leave a blank slot, so
// that the caller may choose what to do.
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

// generate will create a new key and store it in Redis. It relies on a Redis lock to coordinate
// with other AuthN servers.
func (m *maintainer) generate() (*rsa.PrivateKey, error) {
	bucket := m.currentBucket()
	redisKey := fmt.Sprintf("rsa:%d", bucket)
	for {
		// check if another server has created it
		existingKey, err := m.find(bucket)
		if err != nil {
			return nil, err
		}
		if existingKey != nil {
			return existingKey, nil
		}
		// acquire Redis lock (global mutex)
		success, err := m.client.SetNX(redisKey, placeholder, m.race).Result()
		if err != nil {
			return nil, err
		}
		if success {
			// create a new key
			key, err := rsa.GenerateKey(rand.Reader, m.keyStrength)
			if err != nil {
				return nil, err
			}
			// encrypt the key
			ciphertext, err := compat.Encrypt(keyToBytes(key), m.encryptionKey)
			if err != nil {
				return nil, err
			}
			// store the key
			err = m.client.Set(redisKey, ciphertext, m.interval*2+10*time.Second).Err()
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

// find will retrieve and deserialize/decrypt from Redis
func (m *maintainer) find(bucket int64) (*rsa.PrivateKey, error) {
	blob, err := m.client.Get(fmt.Sprintf("rsa:%d", bucket)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "Get")
	} else if blob == placeholder {
		return nil, nil
	}

	plaintext, err := compat.Decrypt([]byte(blob), m.encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "Decrypt")
	}

	return bytesToKey([]byte(plaintext)), nil
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
