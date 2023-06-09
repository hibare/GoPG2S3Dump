package pgdump

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hibare/GoPG2S3Dump/internal/config"
	"github.com/hibare/GoPG2S3Dump/internal/constants"
	"github.com/hibare/GoPG2S3Dump/internal/file"
	"github.com/hibare/GoPG2S3Dump/internal/s3"
	"github.com/hibare/GoPG2S3Dump/internal/utils"
)

var BackupLocation = filepath.Join(os.TempDir(), constants.ExportDir)

func setPGEnvVars() {
	os.Setenv("PGUSER", config.Current.Postgres.User)
	os.Setenv("PGPASSWORD", config.Current.Postgres.Password)
	os.Setenv("PGHOST", config.Current.Postgres.Host)
	os.Setenv("PGPORT", config.Current.Postgres.Port)
}

func createBackupDir() error {
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

func Dump() (int, string, error) {
	var totalDatabases int
	var key string

	setPGEnvVars()

	if err := createBackupDir(); err != nil {
		return totalDatabases, key, err
	}

	// Get list of all databases
	cmd := exec.Command("psql", "-l", "-t")
	output, err := cmd.Output()
	if err != nil {
		return totalDatabases, key, fmt.Errorf("error getting list of databases: %s", err)
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
	totalDatabases = len(databases)

	log.Infof("Exporting %d databases to %s", totalDatabases, BackupLocation)
	// Loop through databases and dump each one to a .sql file
	for _, db := range databases {
		log.Infof("Processing database: %s", db)

		// Dump database to .sql file
		cmd := exec.Command("pg_dump", "--no-owner", "--no-acl", "--dbname="+db, "--file="+db+".sql")
		if err := cmd.Run(); err != nil {
			log.Warnf("Error dumping database: %s, %s", db, err)
			continue
		}
		log.Infof("Successfully dumped database: %s", db)
	}

	log.Infof("Exported %d databases", totalDatabases)

	archivePath, err := file.ArchiveDir(BackupLocation)
	if err != nil {
		return totalDatabases, key, err
	}

	key, err = s3.Upload(archivePath)
	if err != nil {
		return totalDatabases, key, err
	}

	if err = os.Remove(archivePath); err != nil {
		return totalDatabases, key, err
	}
	log.Infof("Removed archive file %s", archivePath)

	return totalDatabases, key, nil
}

func ListDumps() ([]string, error) {
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
	sortedKeys := utils.SortDateTimes(keys)

	return sortedKeys, nil
}

func PurgeDumps() error {
	prefix := s3.GetPrefix()
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
		key = filepath.Join(prefix, key)

		if err := s3.DeleteObjects(key, true); err != nil {
			log.Errorf("Error deleting backup %s: %v", key, err)
			return fmt.Errorf("error deleting backup %s: %v", key, err)
		}
	}

	log.Info("Deletion completed successfully")
	return nil
}
