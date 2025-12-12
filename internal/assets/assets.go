package assets

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chai2010/webp"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"github.com/jaysongiroux/mdserve/internal/routines"
	"golang.org/x/sync/errgroup"
)

var (
	NonOptimizableImageRegexes = []*regexp.Regexp{
		regexp.MustCompile(`apple-touch-icon.[A-z]*`),
		regexp.MustCompile(`favicon-[0-9]*[A-z][0-9]*.[A-z]*`),
	}
)

func OptimizeAssets(
	assets []string,
	siteManifestIconPaths []string,
	serverConfig *config.ServerConfig,
) error {
	if len(assets) == 0 {
		logger.Debug("No assets to optimize")
		return nil
	}

	// Calculate optimal number of workers
	maxWorkers := routines.CalculateMaxWorkers(len(assets))
	logger.Info("Optimizing %d assets with %d concurrent workers", len(assets), maxWorkers)

	// Create error group with limit
	g := new(errgroup.Group)
	g.SetLimit(maxWorkers)

	// Process each asset concurrently
	for i, asset := range assets {
		g.Go(func() error {
			logger.Info("[Worker %d] Checking asset: %s", i, asset)

			// Filter out assets that should be skipped
			if shouldSkipAsset(asset, siteManifestIconPaths) {
				logger.Info(
					"[Worker %d] Skipping asset: %s because it is a site manifest icon, favicon, SVG, or a format that is not supported for optimization",
					i,
					asset,
				)
				return nil
			}

			// Convert to WebP
			if err := convertToWebP(
				asset,
				serverConfig.OptimizeImages,
				serverConfig.OptimizeImagesQuality,
			); err != nil {
				logger.Error("[Worker %d] Failed to convert asset to WebP: %s - %v", i, asset, err)
				return fmt.Errorf("failed to convert asset %s to WebP: %w", asset, err)
			}

			// Delete original asset
			if err := DeleteAsset(asset); err != nil {
				logger.Error("[Worker %d] Failed to delete asset: %s - %v", i, asset, err)
				return fmt.Errorf("failed to delete asset %s: %w", asset, err)
			}

			logger.Debug("[Worker %d] Successfully optimized asset: %s", i, asset)
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		logger.Fatal("[Worker %d] Asset optimization failed: %v", err)
		return err
	}

	logger.Info("Successfully optimized all assets")
	return nil
}

// convertToWebP reads an image and saves a .webp version next to it
func convertToWebP(path string, optimize bool, quality int) error {
	// 1. Open original file
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Error("Failed to close file: %v", err)
		}
	}()

	// 2. Decode (Auto-detects format via imports)
	img, _, err := image.Decode(file)
	if err != nil {
		// If decode fails (e.g. corrupted image), we log and skip
		logger.Error("Skipping %s: could not decode", path)
		return err
	}

	// 3. Create Output File (path/image.jpg -> path/image.webp)
	ext := filepath.Ext(path)
	webpPath := strings.TrimSuffix(path, ext) + ".webp"

	outFile, err := os.Create(filepath.Clean(webpPath))
	if err != nil {
		return err
	}
	defer func() {
		err := outFile.Close()
		if err != nil {
			logger.Error("Failed to close output file: %v", err)
		}
	}()

	// 4. Encode as WebP
	// Lossless: false = similar to JPEG compression
	// Quality: 80 = Good balance for web
	options := webp.Options{
		Lossless: !optimize,
		Quality:  float32(quality),
	}
	if err := webp.Encode(outFile, img, &options); err != nil {
		return err
	}

	logger.Info("Generated WebP: %s", webpPath)
	return nil
}

type SiteWebmanifest struct {
	Icons []Icon `json:"icons"`
}

type Icon struct {
	Src string `json:"src"`
}

func GetIconPathsFromSiteWebmanifest(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	webmanifest, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var siteWebmanifest SiteWebmanifest
	err = json.Unmarshal(webmanifest, &siteWebmanifest)
	if err != nil {
		return nil, err
	}

	var iconPaths []string
	for _, icon := range siteWebmanifest.Icons {
		// remove any leading slash from the icon path
		icon.Src = strings.TrimPrefix(icon.Src, "/")
		// remove any trailing slash from the icon path
		icon.Src = strings.TrimSuffix(icon.Src, "/")
		iconPaths = append(iconPaths, icon.Src)
	}

	return iconPaths, nil
}

func shouldSkipAsset(asset string, siteManifestIconPaths []string) bool {
	// if any of the site manifest icon paths are a suffix of the asset path, return true
	for _, iconPath := range siteManifestIconPaths {
		if strings.HasSuffix(asset, iconPath) {
			return true
		}
	}

	// skip assets that end in .ico
	if strings.HasSuffix(asset, ".ico") {
		return true
	}

	// skip assets that svg
	if strings.HasSuffix(asset, ".svg") {
		return true
	}

	// skip assets that match any of the non-optimizable image regexes
	for _, regex := range NonOptimizableImageRegexes {
		if regex.MatchString(asset) {
			return true
		}
	}

	return false
}
