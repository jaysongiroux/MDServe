package assets

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/otiai10/copy"
)

// GetAssets scans a directory and returns all file paths
func GetAssets(assetsPath string) ([]string, error) {
	files, err := os.ReadDir(assetsPath)
	if err != nil {
		return nil, err
	}

	var filePaths []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePaths = append(filePaths, filepath.Join(assetsPath, file.Name()))
	}

	return filePaths, nil
}

// GetOptimizableAssets finds standard images (JPG/PNG/GIF)
func GetOptimizableAssets(paths []string) ([]string, error) {
	var optimizableAssets []string
	for _, path := range paths {
		// We skip files that are already WebP
		if slices.Contains(constants.OptimizableImageExtensions, strings.ToLower(filepath.Ext(path))) {
			optimizableAssets = append(optimizableAssets, path)
		}
	}
	return optimizableAssets, nil
}

// copies all files and folders from the source path to the destination path
func CopyAssets(sourcePath string, destinationPath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip the source directory itself
		if path == sourcePath {
			return nil
		}

		// exclude all files and folders that start with a dot
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// exclude all files and folders that are in the .generated-assets folder
		if strings.HasPrefix(info.Name(), constants.GeneratedAssetsPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// get the relative path from source
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		// copy the file or folder to the destination path
		destPath := filepath.Join(destinationPath, relPath)
		err = copy.Copy(path, destPath)
		if err != nil {
			return err
		}

		return nil
	})
}

func MoveAssets(assetsPath string, destinationPath string) error {
	// copy all files from the assets path to the destination path
	err := CopyAssets(assetsPath, destinationPath)
	if err != nil {
		return err
	}

	return nil
}

func DeleteAsset(path string) error {
	return os.Remove(path)
}
