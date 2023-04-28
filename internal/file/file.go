package file

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func ArchiveDir(dirPath string) (string, error) {
	dirPath = filepath.Clean(dirPath)
	dirName := filepath.Base(dirPath)
	zipName := fmt.Sprintf("%s.zip", dirName)
	zipPath := filepath.Join(os.TempDir(), zipName)

	// Create a temporary file to hold the zip archive
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return zipPath, err
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			log.Errorf("Failed to get file info: %v", err)
			return nil
		}

		if !info.Mode().IsRegular() {
			log.Warnf("%s is not a regular file", path)
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			log.Errorf("Failed to get relative path: %v", err)
			return nil
		}
		header.Name = filepath.ToSlash(filepath.Join(relPath))

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Errorf("Failed to open file: %v", err)
			return nil
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			file.Close()
			log.Errorf("Failed to write file to archive: %v", err)
			return nil
		}
		file.Close()

		return nil
	})

	log.Infof("Created archive '%s' for directory '%s'", zipPath, dirPath)
	return zipPath, err
}
