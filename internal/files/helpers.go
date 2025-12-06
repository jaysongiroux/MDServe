package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func DeleteDirectory(path string) error {
	return os.RemoveAll(path)
}

func GetFileModifiedDate(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func CopyFile(sourcePath, destinationPath string, overwrite bool) error {
	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Check destination existence
	if info, err := os.Stat(destinationPath); err == nil {
		if info.IsDir() {
			return fmt.Errorf("destination path is a directory: %s", destinationPath)
		}
		if !overwrite {
			return nil
		}
	} else if !os.IsNotExist(err) {
		// Real error (permissions, IO, etc.)
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return err
	}

	// Create/overwrite destination file
	dst, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Perform the copy
	_, err = io.Copy(dst, src)
	return err
}
