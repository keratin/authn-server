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
func (m *maintainer) maintain(ks *keyStore) error {
	// restore current keys, if any
	keys, err := m.restore()
	if err != nil {
		return errors.Wrap(err, "restore")
	}
	for _, key := range keys {
		ks.Rotate(key)
	}

	// ensure at least one key (cold start)
	// BUG: when the *previous* key exists but the *current* key does not, we still need to generate
	if len(keys) == 0 {
		newKey, err := m.generate()
		if err != nil {
			return errors.Wrap(err, "generate")
		}
		ks.Rotate(newKey)
	}

	go func() {
		// sleep until next interval change
		elapsedSeconds := time.Now().Unix() % int64(m.interval/time.Second)
		time.Sleep(m.interval - time.Duration(elapsedSeconds)*time.Second)

		// rotate at regular intervals
		ticker := time.NewTicker(m.interval)
		m.rotate(ks)
		for {
			<-ticker.C
			m.rotate(ks)
		}
	}()

	return nil
}

func (m *maintainer) rotate(ks *keyStore) {
	newKey, err := m.generate()
	if err != nil {
		// TODO: report
		return
	}
	ks.Rotate(newKey)
}

// restore will query Redis for the previous and current keys. It returns keys in the proper sorting
// order, with the newest (current) key in last position.
func (m *maintainer) restore() ([]*rsa.PrivateKey, error) {
	bucket := m.currentBucket()
	keys := []*rsa.PrivateKey{}

	previous, err := m.find(bucket - 1)
	if err != nil {
		return nil, err
	}
	if previous != nil {
		keys = append(keys, previous)
	}

	current, err := m.find(bucket)
	if err != nil {
		return nil, err
	}
	if current != nil {
		keys = append(keys, current)
	}

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
		return nil, errors.Wrap(err, "Redis Get")
	} else if blob == placeholder {
		return nil, nil
	}

	plaintext, err := compat.Decrypt([]byte(blob), m.encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "compat.Decrypt")
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
