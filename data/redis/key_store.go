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
	keys          map[int64]*rsa.PrivateKey
	localMutex    sync.Mutex
}

// NewKeyStore creates a key store that uses Redis to persist an auto-generated key and rotate it
// regularly. The key is encrypted using SECRET_KEY_BASE, which is already the ultimate SPOF for
// AuthN security. It's expected that very few people will be in position to improve on the security
// tradeoffs of this provider.
func NewKeyStore(client *redis.Client, interval time.Duration, race time.Duration, encryptionKey []byte) *keyStore {
	return &keyStore{
		interval:      interval,
		race:          race,
		keyStrength:   2048,
		client:        client,
		encryptionKey: encryptionKey,
		localMutex:    sync.Mutex{},
		keys:          make(map[int64]*rsa.PrivateKey),
	}
}

func (ks *keyStore) Key() (*rsa.PrivateKey, error) {
	bucket := ks.currentBucket()

	if ks.keys[bucket] == nil {
		// this lock only prevents stampedes within a single server. it does
		// not prevent stampedes across multiple servers, and it does not make
		// the keys map safe for concurrency.
		ks.localMutex.Lock()
		defer ks.localMutex.Unlock()

		// if we were waiting, another routine may have already done it.
		if ks.keys[bucket] == nil {
			key, err := ks.findOrCreate(bucket)
			if err != nil {
				return nil, err
			}
			ks.keys[bucket] = key
		}

		// trim out old keys, keeping only the current and previous
		go func() {
			for b := range ks.keys {
				if b+1 < bucket {
					delete(ks.keys, b)
				}
			}
		}()
	}

	return ks.keys[bucket], nil
}

// Keys must reliably return the keys associated with the current interval
// and the previous interval. Intervals are not guaranteed to have a key,
// if no signing took place in that interval. The local cache of keys is
// also not guaranteed to exist or have the appropriate contents, since this
// server may not have been asked to sign or report on keys since starting.
func (ks *keyStore) Keys() ([]*rsa.PrivateKey, error) {
	bucket := ks.currentBucket()
	keys := []*rsa.PrivateKey{}

	previous, err := ks.findAndRemember(bucket - 1)
	if err != nil {
		return nil, err
	}
	if previous != nil {
		keys = append(keys, previous)
	}

	current, err := ks.findAndRemember(bucket)
	if err != nil {
		return nil, err
	}
	if current != nil {
		keys = append(keys, current)
	}

	return keys, nil
}

// find will retrieve and deserialize/decrypt from Redis
func (ks *keyStore) find(bucket int64) (*rsa.PrivateKey, error) {
	blob, err := ks.client.Get(fmt.Sprintf("rsa:%d", bucket)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else if blob == placeholder {
		return nil, nil
	}

	plaintext, err := compat.Decrypt([]byte(blob), ks.encryptionKey)
	if err != nil {
		return nil, err
	}

	return bytesToKey([]byte(plaintext)), nil
}

// findAndRemember will find a key from Redis and memoize the result
func (ks *keyStore) findAndRemember(bucket int64) (*rsa.PrivateKey, error) {
	// maybe previously remembered
	if ks.keys[bucket] != nil {
		return ks.keys[bucket], nil
	}
	// fetch
	key, err := ks.find(bucket)
	if err != nil || key == nil {
		return key, err
	}
	// remember
	ks.keys[bucket] = key
	return ks.keys[bucket], nil
}

// findOrCreate will find a key from Redis or acquire a global lock before
// creating a missing key and storing it in Redis.
func (ks *keyStore) findOrCreate(bucket int64) (*rsa.PrivateKey, error) {
	fmt.Println(bucket)
	redisKey := fmt.Sprintf("rsa:%d", bucket)
	for {
		fmt.Println("loop")
		// check if it exists (not a placeholder)
		val, err := ks.find(bucket)
		if err != nil || val != nil {
			return val, err
		}
		// attempt to get redis lock (global mutex with other servers)
		success, err := ks.client.SetNX(redisKey, placeholder, ks.race).Result()
		if err != nil {
			return nil, err
		}
		if success {
			// create a new key
			key, err := rsa.GenerateKey(rand.Reader, ks.keyStrength)
			if err != nil {
				return nil, err
			}
			// encrypt the key
			ciphertext, err := compat.Encrypt(keyToBytes(key), ks.encryptionKey)
			if err != nil {
				return nil, err
			}
			// store the key
			err = ks.client.Set(redisKey, ciphertext, ks.interval*2+10*time.Second).Err()
			if err != nil {
				return nil, err
			}
			// return the key
			return key, nil
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func (ks *keyStore) currentBucket() int64 {
	return time.Now().Unix() / int64(ks.interval)
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
