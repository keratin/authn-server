package data

type BlobStore interface {
	// Read fetches a blob from the store.
	Read(name string) ([]byte, error)

	// WLock acquires a global mutex that will either timeout or be
	// released by a successful Write
	WLock(name string) (bool, error)

	// Write puts a blob into the store.
	Write(name string, blob []byte) error
}
