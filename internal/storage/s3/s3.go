// Package s3 provides an implementation of storage interface for S3-compatible backends.
package s3

import (
	"fmt"
	"path/filepath"

	commonS3 "github.com/hibare/GoCommon/v2/pkg/s3"
	"github.com/hibare/stashly/internal/config"
)

// S3 implements the StorageIface for S3-compatible storage backends.
type S3 struct {
	s3  *commonS3.S3
	cfg *config.Config
}

// NewS3Storage creates a new S3Storage instance with the provided configuration.
func NewS3Storage(cfg *config.Config) *S3 {
	s := &S3{
		s3: &commonS3.S3{
			Endpoint:  cfg.S3.Endpoint,
			Region:    cfg.S3.Region,
			AccessKey: cfg.S3.AccessKey,
			SecretKey: cfg.S3.SecretKey,
			Bucket:    cfg.S3.Bucket,
		},
		cfg: cfg,
	}
	// Set prefix (hostname included)
	s.s3.SetPrefix(cfg.S3.Prefix, cfg.App.InstanceID, true)
	return s
}

// Init prepares the S3 storage by establishing a session.
func (s *S3) Init() error {
	if err := s.s3.NewSession(); err != nil {
		return err
	}
	return nil
}

// Name returns the name of the storage backend (e.g., "s3").
func (s *S3) Name() string {
	return fmt.Sprintf("s3 (%s)", s.s3.Bucket)
}

// Upload uploads a local file to S3 and returns the remote key/path.
func (s *S3) Upload(localPath string) (string, error) {
	key, err := s.s3.UploadFile(localPath)
	if err != nil {
		return "", err
	}
	return key, nil
}

// List returns keys/identifiers under the configured prefix.
func (s *S3) List() ([]string, error) {
	s.s3.SetPrefix(s.cfg.S3.Prefix, s.cfg.App.InstanceID, false)
	if err := s.s3.NewSession(); err != nil {
		return nil, err
	}
	keys, err := s.s3.ListObjectsAtPrefixRoot()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Delete deletes the provided key/path from S3 storage.
func (s *S3) Delete(key string) error {
	// key may be datetime string - join with prefix
	fullKey := filepath.Join(s.s3.Prefix, key)
	return s.s3.DeleteObjects(fullKey, true)
}

// TrimPrefix trims the configured prefix from a given key, if present.
func (s *S3) TrimPrefix(keys []string) []string {
	return s.s3.TrimPrefix(keys)
}
