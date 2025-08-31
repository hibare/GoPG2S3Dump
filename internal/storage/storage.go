// Package storage defines the interface for various storage backends.
package storage

// StorageIface defines a generic storage backend used to upload and manage backups.
// revive:disable-next-line exported
type StorageIface interface {
	// Init prepares the storage (e.g., establishes session)
	Init() error

	// Name returns the name of the storage backend (e.g., "s3", "gcs")
	Name() string

	// Upload uploads a local file and returns the remote key/path
	Upload(localPath string) (string, error)

	// List returns keys/identifiers under configured prefix
	List() ([]string, error)

	// Delete deletes the provided key/path from storage
	Delete(key string) error

	// TrimPrefix trims the configured prefix from a given key, if present
	TrimPrefix(keys []string) []string
}
