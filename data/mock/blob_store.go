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

func (bs *BlobStore) Write(name string, blob []byte) error {
	bs.blobs[name] = blob
	return nil
}

func (bs *BlobStore) Read(name string) ([]byte, error) {
	val := bs.blobs[name]
	if string(val) == placeholder {
		return nil, nil
	}
	return val, nil
}

func (bs *BlobStore) WLock(name string) (bool, error) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if bs.blobs[name] == nil {
		bs.blobs[name] = []byte(placeholder)
		return true, nil
	}
	return false, nil
}
