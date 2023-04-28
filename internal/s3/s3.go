package s3

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hibare/GoPG2S3Dump/internal/config"
	"github.com/hibare/GoPG2S3Dump/internal/constants"
)

func GetPrefix() string {
	prefixSlice := []string{}
	prefix := config.Current.S3.Prefix
	hostname := config.Current.Backup.Hostname

	if prefix != "" {
		prefixSlice = append(prefixSlice, prefix)
	}

	if hostname != "" {
		prefixSlice = append(prefixSlice, hostname)
	}

	generatedPrefix := filepath.Join(prefixSlice...)

	if !strings.HasSuffix(generatedPrefix, constants.PrefixSeparator) {
		generatedPrefix = fmt.Sprintf("%s%s", generatedPrefix, constants.PrefixSeparator)
	}

	return generatedPrefix
}

func GetTimeStampedPrefix() string {
	timePrefix := time.Now().Format(constants.DefaultDateTimeLayout)
	prefix := GetPrefix()

	generatedPrefix := filepath.Join(prefix, timePrefix)

	if !strings.HasSuffix(generatedPrefix, constants.PrefixSeparator) {
		generatedPrefix = fmt.Sprintf("%s%s", generatedPrefix, constants.PrefixSeparator)
	}

	return generatedPrefix

}

func TrimPrefix(keys []string) []string {
	var trimmedKeys []string
	prefix := GetPrefix()

	for _, key := range keys {
		trimmedKey := strings.TrimPrefix(key, prefix)
		trimmedKey = strings.TrimSuffix(trimmedKey, "/")
		trimmedKeys = append(trimmedKeys, trimmedKey)
	}
	return trimmedKeys
}

func NewSession() (*session.Session, error) {
	s3Config := config.Current.S3

	sess, err := session.NewSession(&aws.Config{
		Region:      &s3Config.Region,
		Endpoint:    &s3Config.Endpoint,
		Credentials: credentials.NewStaticCredentials(s3Config.AccessKey, s3Config.SecretKey, ""),
	})

	if err != nil {
		return nil, fmt.Errorf("error creating session: %v", err)
	}

	return sess, nil
}

func Upload(archivePath string) (string, error) {
	bucket := config.Current.S3.Bucket
	prefix := GetTimeStampedPrefix()
	key := filepath.Join(prefix, filepath.Base(archivePath))

	// Create a session
	sess, err := NewSession()

	if err != nil {
		return key, err
	}

	// Create an S3 uploader using the session
	uploader := s3manager.NewUploader(sess)

	// Open the file to upload
	f, err := os.Open(archivePath)
	if err != nil {
		log.Errorf("Failed to open file %v", err)
		return key, err
	}
	defer f.Close()

	log.Infof("Uploading file %s to S3://%s/%s", archivePath, bucket, prefix)

	// Upload the file to S3
	if _, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	}); err != nil {
		log.Errorf("Failed to upload file %v", err)
		return key, err
	}
	log.Infof("Uploaded %s to S3://%s/%s", archivePath, bucket, key)

	return key, nil
}

func ListObjectsAtPrefixRoot() ([]string, error) {
	var keys []string
	prefix := GetPrefix()

	sess, err := NewSession()
	if err != nil {
		return keys, err
	}
	client := s3.New(sess)

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(config.Current.S3.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	resp, err := client.ListObjectsV2(input)
	if err != nil {
		return keys, err
	}

	for _, obj := range resp.Contents {
		if *obj.Key == prefix {
			continue
		}
		keys = append(keys, *obj.Key)
	}

	if len(keys) == 0 && len(resp.CommonPrefixes) == 0 {
		return keys, nil
	}

	for _, cp := range resp.CommonPrefixes {
		keys = append(keys, *cp.Prefix)
	}

	return keys, nil
}

func DeleteObjects(key string, recursive bool) error {
	bucket := config.Current.S3.Bucket

	sess, err := NewSession()
	if err != nil {
		return err
	}
	client := s3.New(sess)

	// Delete all child object recursively
	if recursive {
		log.Warnf("Recursively deleting objects in bucket S3://%s/%s", bucket, key)
		// List all objects in the bucket with the given key
		resp, err := client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(key),
		})
		if err != nil {
			return err
		}

		log.Infof("Found %d objects in bucket S3://%s/%s", len(resp.Contents), bucket, key)

		// Delete all objects with the given key
		for _, obj := range resp.Contents {
			_, err = client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    obj.Key,
			})

			if err != nil {
				return err
			}
			log.Infof("Deleted object with key '%s' from bucket '%s'", *obj.Key, bucket)
		}
	}

	// Delete the prefix
	if _, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		return err
	}

	log.Infof("Deleted key %s", key)

	return nil
}
