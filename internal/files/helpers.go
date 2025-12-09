package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jaysongiroux/mdserve/internal/logger"
)

func GetFileModifiedDate(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func CheckIfDirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetAllFilesInDirectory(path string) ([]string, error) {
	var filePaths []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only add files, not directories
		if !info.IsDir() {
			filePaths = append(filePaths, filePath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

func RecursivelyCopyDirectory(sourcePath string, destinationPath string) error {
	// Get source info to preserve permissions
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory with source permissions
	logger.Debug("Creating destination directory %s with source permissions %v", destinationPath, sourceInfo.Mode())
	if err := os.MkdirAll(destinationPath, sourceInfo.Mode()); err != nil {
		logger.Error("Failed to create destination directory %s with source permissions %v: %v", destinationPath, sourceInfo.Mode(), err)
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		logger.Debug("Walking source path %s", path)
		if err != nil {
			logger.Error("Failed to walk source path %s: %v", sourcePath, err)
			return err
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			logger.Error("Failed to get relative path from source path %s to path %s: %v", sourcePath, path, err)
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Calculate destination path
		destPath := filepath.Join(destinationPath, relPath)

		// Handle directories
		if info.IsDir() {
			// Create directory with same permissions
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				logger.Error("Failed to create directory %s with source permissions %v: %v", destPath, info.Mode(), err)
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			return nil
		}

		// Handle files
		return CopyFile(path, destPath, true)
	})
}

func CopyFile(sourcePath, destinationPath string, overwrite bool) error {
	// Get source file info
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Check if source is a directory
	if sourceInfo.IsDir() {
		return fmt.Errorf("source path is a directory, use RecursivelyCopyDirectory: %s", sourcePath)
	}

	// Check if destination exists
	destInfo, err := os.Stat(destinationPath)
	if err == nil {
		// Destination exists
		if destInfo.IsDir() {
			return fmt.Errorf("destination path is a directory: %s", destinationPath)
		}
		if !overwrite {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat destination: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file with source permissions
	dst, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Perform the copy
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure data is flushed to disk
	if err := dst.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

func DeleteDirectoryContents(directoryPath string) error {
	// Check if the directory exists
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		logger.Debug("Directory %s does not exist", directoryPath)
		return nil
	}

	logger.Debug("Deleting directory contents of %s", directoryPath)

	// Read all entries in the directory
	entries, err := os.ReadDir(directoryPath)
	if err != nil {
		logger.Error("Failed to read directory %s: %v", directoryPath, err)
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Delete each entry
	for _, entry := range entries {
		entryPath := filepath.Join(directoryPath, entry.Name())

		if entry.IsDir() {
			logger.Debug("Deleting subdirectory %s", entryPath)
			if err := os.RemoveAll(entryPath); err != nil {
				logger.Error("Failed to delete subdirectory %s: %v", entryPath, err)
				return fmt.Errorf("failed to delete subdirectory %s: %w", entryPath, err)
			}
		} else {
			logger.Debug("Deleting file %s", entryPath)
			if err := os.Remove(entryPath); err != nil {
				logger.Error("Failed to delete file %s: %v", entryPath, err)
				return fmt.Errorf("failed to delete file %s: %w", entryPath, err)
			}
		}
	}

	logger.Debug("Successfully deleted directory contents of %s", directoryPath)
	return nil
}
