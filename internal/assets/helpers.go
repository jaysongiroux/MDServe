package assets

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/files"
)

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

func MoveAssets(assetsPath string, destinationPath string) error {
	// copy all files from the assets path to the destination path
	err := files.RecursivelyCopyDirectory(assetsPath, destinationPath)
	if err != nil {
		return err
	}

	return nil
}

func DeleteAsset(path string) error {
	return os.Remove(path)
}
