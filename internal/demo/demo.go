package demo

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/handler"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

func FetchReadme() (*string, error) {
	response, err := http.Get(constants.DemoReadmeURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	readme := string(body)
	return &readme, nil
}

func HandleDemoEnabled(app *handler.App) error {
	readme, err := FetchReadme()
	if err != nil {
		return err
	}

	fmt.Println("readme", *readme)

	logger.Info("README fetched: %s", *readme)

	return fmt.Errorf("demo mode is not implemented")
	// // if demo is true, copy the README to the content folder as index.md
	// // if index.md already exists, log a warning and do nothing
	// if app.ServerConfig.Demo {
	// 	app.Logger.Warn("WARNING: Demo mode is enabled. Copying README to content folder as index.md")

	// 	readmePath := filepath.Join(app.ServerConfig.ContentPath, "index.md")
	// 	if _, err := os.Stat(readmePath); err == nil {
	// 		app.Logger.Warn("README already exists at %s", readmePath)
	// 	} else {
	// 		repoReadmePath := "README.md"
	// 		err = files.CopyFile(repoReadmePath, readmePath, true)
	// 		if err != nil {
	// 			app.Logger.Error("Failed to copy README: %v", err)
	// 			return err
	// 		}
	// 		app.Logger.Info("README copied to %s", readmePath)
	// 	}
	// }
}
