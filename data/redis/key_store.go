package redis

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/compat"
)

var placeholder = "generating"

type keyStore struct {
	keys   []*rsa.PrivateKey
	rwLock *sync.RWMutex
}

// NewKeyStore creates a key store that uses Redis to persist an auto-generated key and rotate it
// regularly. The key is encrypted using SECRET_KEY_BASE, which is already the ultimate SPOF for
// AuthN security. It's expected that very few people will be in position to improve on the security
// tradeoffs of this provider.
func NewKeyStore(client *redis.Client, interval time.Duration, race time.Duration, encryptionKey []byte) (*keyStore, error) {
	ks := &keyStore{
		keys:   []*rsa.PrivateKey{},
		rwLock: &sync.RWMutex{},
	}

	m := &maintainer{
		interval:      interval,
		race:          race,
		keyStrength:   2048,
		client:        client,
		encryptionKey: encryptionKey,
	}
	err := m.maintain(ks)
	if err != nil {
		return nil, err
	}

	return ks, nil
}

// Key returns the current key. It relies on the internal keys slice being sorted with the newest
// key last.
func (ks *keyStore) Key() *rsa.PrivateKey {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	return ks.keys[len(ks.keys)-1]
}

// Keys will return the previous and current keys, in that order.
func (ks *keyStore) Keys() []*rsa.PrivateKey {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	return ks.keys
}

// Rotate is responsible for adding a new key to the list. It maintains key order from oldest to
// newest, and ensures a maximum of two entries.
func (ks *keyStore) Rotate(k *rsa.PrivateKey) {
	keys := []*rsa.PrivateKey{}
	if len(ks.keys) > 0 {
		keys = append(keys, ks.keys[len(ks.keys)-1])
	}
	keys = append(keys, k)

	ks.rwLock.Lock()
	defer ks.rwLock.Unlock()
	ks.keys = keys
}

//-- MAINTAINER

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
		return err
	}
	for _, key := range keys {
		ks.Rotate(key)
	}

	// ensure at least one key (cold start)
	if len(keys) == 0 {
		newKey, err := m.generate()
		if err != nil {
			return err
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
		// TODO: report and continue
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
		return nil, err
	} else if blob == placeholder {
		return nil, nil
	}

	plaintext, err := compat.Decrypt([]byte(blob), m.encryptionKey)
	if err != nil {
		return nil, err
	}

	return bytesToKey([]byte(plaintext)), nil
}

func (m *maintainer) currentBucket() int64 {
	return time.Now().Unix() / int64(m.interval/time.Second)
}

//-- UTIL

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
