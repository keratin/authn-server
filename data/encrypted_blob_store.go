package data

import "github.com/keratin/authn-server/lib/compat"

type EncryptedBlobStore struct {
	store         BlobStore
	encryptionKey []byte
}

func NewEncryptedBlobStore(store BlobStore, encryptionKey []byte) *EncryptedBlobStore {
	return &EncryptedBlobStore{
		store:         store,
		encryptionKey: encryptionKey,
	}
}

func (bs *EncryptedBlobStore) Read(name string) ([]byte, error) {
	encryptedBlob, err := bs.store.Read(name)
	if err != nil || encryptedBlob == nil {
		return encryptedBlob, err
	}
	val, err := compat.Decrypt(encryptedBlob, bs.encryptionKey)
	return []byte(val), err
}

func (bs *EncryptedBlobStore) WriteNX(name string, blob []byte) (bool, error) {
	encryptedBlob, err := compat.Encrypt(blob, bs.encryptionKey)
	if err != nil {
		return false, err
	}
	return bs.store.WriteNX(name, encryptedBlob)
}
