package assets

import (
	"image"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

// convertToWebP reads an image and saves a .webp version next to it
func ConvertToWebP(path string, optimize bool, quality int) error {
	// 1. Open original file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 2. Decode (Auto-detects format via imports)
	img, _, err := image.Decode(file)
	if err != nil {
		// If decode fails (e.g. corrupted image), we log and skip
		log.Printf("Skipping %s: could not decode", path)
		return err
	}

	// 3. Create Output File (path/image.jpg -> path/image.webp)
	ext := filepath.Ext(path)
	webpPath := strings.TrimSuffix(path, ext) + ".webp"

	outFile, err := os.Create(webpPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

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
