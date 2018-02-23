package mock

import "time"
import "sync"

type BlobStore struct {
	blobs    map[string][]byte
	mutex    sync.Mutex
	TTL      time.Duration
	LockTime time.Duration
}

var placeholder = "mock-blob-store"

func NewBlobStore(ttl time.Duration, lockTime time.Duration) *BlobStore {
	return &BlobStore{
		blobs:    map[string][]byte{},
		mutex:    sync.Mutex{},
		TTL:      ttl,
		LockTime: lockTime,
	}
}

func (bs *BlobStore) Read(name string) ([]byte, error) {
	val := bs.blobs[name]
	if string(val) == placeholder {
		return nil, nil
	}
	return val, nil
}

func (bs *BlobStore) WriteNX(name string, blob []byte) (bool, error) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if bs.blobs[name] != nil {
		return false, nil
	}
	bs.blobs[name] = blob
	return true, nil
}
