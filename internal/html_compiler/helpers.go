package htmlcompiler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/constants"
)

// gets a list of all mark down files in the content directory
func GetMDFiles(mdFilesPath string) ([]string, error) {
	var filePaths []string

	err := filepath.Walk(mdFilesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".md" {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

func WriteHTMLFile(basePath string, fileName string, html string) error {
	// check if the HTMLFilesPath exists
	htmlFilesPath := filepath.Join(basePath, constants.HTMLFilesPath)
	if _, err := os.Stat(htmlFilesPath); os.IsNotExist(err) {
		err = os.MkdirAll(htmlFilesPath, 0755)
		if err != nil {
			return err
		}
	}

	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	filePath := filepath.Join(htmlFilesPath, fileName+".html")

	// Ensure the directory for the file exists
	fileDir := filepath.Dir(filePath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		err = os.MkdirAll(fileDir, 0755)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(filePath, []byte(html), 0644)
	if err != nil {
		return err
	}
	return nil
}
