package demo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/handler"
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

	// check status code
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch README: %d", response.StatusCode)
	}

	return &readme, nil
}

func HandleDemoEnabled(app *handler.App) error {
	// if demo is true, copy the README to the content folder as index.md
	// if index.md already exists, log a warning and do nothing

	readme, err := FetchReadme()
	if err != nil {
		return err
	}

	app.Logger.Warn("WARNING: Demo mode is enabled. Copying README to content folder as index.md")

	// if the index.md file already exists, log a warning, delete the file and replace it with the fetched README
	readmePath := filepath.Join(app.ServerConfig.ContentPath, "index.md")
	if _, err := os.Stat(readmePath); err == nil {
		app.Logger.Warn("README already exists at %s", readmePath)
		err = os.Remove(readmePath)
		if err != nil {
			return err
		}
	}

	// create a comment informing the user that this file was written
	// because the demo mode is enabled
	comment := `<!-- 
	NOTE: 
	This file was written because the demo mode is enabled.
	You can disable this behavior by setting the demo flag to false in the server config file.
	 -->`

	// prepend the comment to the fetched README
	readmeString := *readme
	readmeString = comment + "\n" + readmeString

	// write the fetched README to the content folder as index.md
	err = os.WriteFile(readmePath, []byte(readmeString), 0644)
	if err != nil {
		return err
	}

	app.Logger.Info("README copied to %s", readmePath)

	return nil
}
