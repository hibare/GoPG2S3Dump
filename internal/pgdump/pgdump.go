package pgdump

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hibare/GoCommon/v2/pkg/crypto/gpg"
	"github.com/hibare/GoCommon/v2/pkg/datetime"
	"github.com/hibare/GoCommon/v2/pkg/file"
	commonS3 "github.com/hibare/GoCommon/v2/pkg/s3"
	"github.com/hibare/GoPG2S3Dump/internal/config"
	"github.com/hibare/GoPG2S3Dump/internal/constants"
)

var BackupLocation = filepath.Join(os.TempDir(), constants.ExportDir)

func setPGEnvVars() {
	os.Setenv("PGUSER", config.Current.Postgres.User)
	os.Setenv("PGPASSWORD", config.Current.Postgres.Password)
	os.Setenv("PGHOST", config.Current.Postgres.Host)
	os.Setenv("PGPORT", config.Current.Postgres.Port)
}

func setupBackupDir() error {
	// Remove existing backupLocation
	os.RemoveAll(BackupLocation)

	// Create a new directory in backupLocation
	if err := os.MkdirAll(BackupLocation, 0755); err != nil {
		return fmt.Errorf("error creating directory: %s", err)
	}

	// Change working directory to backupLocation
	if err := os.Chdir(BackupLocation); err != nil {
		return fmt.Errorf("error changing directory: %s", err)
	}

	return nil
}

func exportDBs() (int, error) {
	totalDatabases := 0

	// Get list of all databases
	cmd := exec.Command("psql", "-l", "-t")
	output, err := cmd.Output()
	if err != nil {
		return totalDatabases, fmt.Errorf("error getting list of databases: %s", err)
	}

	databases := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		// Use cut to extract database name
		fields := strings.Split(line, "|")
		if len(fields) > 0 {
			dbName := strings.TrimSpace(fields[0])
			// Use sed to remove whitespace and empty lines
			dbName = strings.TrimSpace(dbName)
			if len(dbName) > 0 && !strings.HasPrefix(dbName, "template") && dbName != "postgres" && dbName != "defaultdb" {
				databases = append(databases, dbName)
			}
		}
	}

	// Loop through databases and dump each one to a .sql file
	for _, db := range databases {
		log.Infof("Processing database: %s", db)

		// Dump database to .sql file
		cmd := exec.Command("pg_dump", "--no-owner", "--no-acl", "--dbname="+db, "--file="+db+".sql")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Warnf("Error dumping database: %s, %s", db, err)
			continue
		}
		totalDatabases = totalDatabases + 1
		log.Infof("Successfully dumped database: %s", db)
	}

	log.Infof("Exported %d databases", totalDatabases)

	return totalDatabases, nil
}

func Dump() (int, string, error) {
	var totalDatabases int
	var archivePath string
	var uploadFilePath string

	setPGEnvVars()

	if err := setupBackupDir(); err != nil {
		return totalDatabases, "", err
	}

	totalDatabases, err := exportDBs()
	if err != nil {
		return totalDatabases, "", err
	}

	if totalDatabases <= 0 {
		return totalDatabases, "", fmt.Errorf("no databases to export")
	}

	// create an archive of dump directory
	archivePath, _, _, _, err = file.ArchiveDir(BackupLocation)
	if err != nil {
		return totalDatabases, "", err
	}

	uploadFilePath = archivePath

	// encrypt backup if encryption is enabled
	if config.Current.Backup.Encrypt {
		gpg, err := gpg.DownloadGPGPubKey(config.Current.Encryption.GPG.KeyID, config.Current.Encryption.GPG.KeyServer)
		if err != nil {
			log.Warnf("Error downloading gpg key: %s", err)
			return totalDatabases, "", err
		}

		encryptedFilePath, err := gpg.EncryptFile(archivePath)
		if err != nil {
			log.Warnf("Error encrypting archive file: %s", err)
			return totalDatabases, "", err
		}

		uploadFilePath = encryptedFilePath
	}

	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}
	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, true)

	if err := s3.NewSession(); err != nil {
		return totalDatabases, "", err
	}

	// upload dump archive to S3
	key, err := s3.UploadFile(uploadFilePath)
	if err != nil {
		return totalDatabases, "", err
	}

	log.Infof("Backup uploaded at %s", key)

	return totalDatabases, key, nil
}

func ListDumps() ([]string, error) {
	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}
	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, false)

	if err := s3.NewSession(); err != nil {
		return []string{}, err
	}

	keys, err := s3.ListObjectsAtPrefixRoot()
	if err != nil {
		return []string{}, err
	}

	if len(keys) == 0 {
		log.Info("No backups found")
		return []string{}, nil
	}

	// Remove prefix from key to get datetime string
	keys = s3.TrimPrefix(keys)

	// Sort datetime strings by descending order
	sortedKeys := datetime.SortDateTimes(keys)

	return sortedKeys, nil
}

func PurgeDumps() error {
	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}
	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, false)

	if err := s3.NewSession(); err != nil {
		return err
	}

	dumps, err := ListDumps()

	if err != nil {
		return err
	}

	if len(dumps) <= int(config.Current.Backup.RetentionCount) {
		log.Info("No backups to delete")
		return nil
	}

	keysToDelete := dumps[config.Current.Backup.RetentionCount:]
	log.Infof("Found %d backups to delete (backup rentention %d) [%s]", len(keysToDelete), config.Current.Backup.RetentionCount, keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		log.Infof("Deleting backup %s", key)
		key = filepath.Join(s3.Prefix, key)

		if err := s3.DeleteObjects(key, true); err != nil {
			log.Errorf("Error deleting backup %s: %v", key, err)
			return fmt.Errorf("error deleting backup %s: %v", key, err)
		}
	}

	log.Info("Deletion completed successfully")
	return nil
}
