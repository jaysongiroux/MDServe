package handler

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/jaysongiroux/mdserve/internal/constants"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

func loadPageContent(app *App, pageName, mdPath string) (template.HTML, error) {
	if app.ServerConfig.HTMLCompilationMode == constants.HTMLCompilationModeStatic {
		return loadStaticHTML(app, pageName)
	}
	return loadAndCompileMarkdown(app, pageName, mdPath)
}

func loadStaticHTML(app *App, pageName string) (template.HTML, error) {
	htmlPath := filepath.Join(
		app.ServerConfig.GeneratedPath,
		constants.HTMLFilesPath,
		pageName+".html",
	)
	contentBytes, err := os.ReadFile(filepath.Clean(htmlPath))

	if os.IsNotExist(err) {
		// Try with /index.html appended
		indexPath := filepath.Join(
			app.ServerConfig.GeneratedPath,
			constants.HTMLFilesPath,
			pageName,
			"index.html",
		)
		contentBytes, err = os.ReadFile(filepath.Clean(indexPath))
		if os.IsNotExist(err) {
			app.Logger.Warn("404 Not Found: %s", htmlPath)
			return "", NewPageError(Err404Code, Err404Title, Err404Message)
		}
	}

	if err != nil {
		app.Logger.Error("Error reading static file: %v", err)
		return "", NewPageError(Err500Code, Err500Title, Err500Message)
	}

	// #nosec G203 -- Content is from trusted markdown files compiled at server startup, not user input
	return template.HTML(contentBytes), nil
}

func loadAndCompileMarkdown(app *App, pageName, mdPath string) (template.HTML, error) {
	// Check if file exists
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		// Try with /index.md appended
		indexPath := filepath.Join(app.ServerConfig.ContentPath, pageName, "index.md")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			app.Logger.Warn("404 Not Found: %s", mdPath)
			return "", NewPageError(Err404Code, Err404Title, Err404Message)
		}
		mdPath = indexPath
	}

	htmlString, err := htmlcompiler.CompileHTMLFile(mdPath, app.SiteConfig)
	if err != nil {
		app.Logger.Error("Error compiling markdown live: %v", err)
		return "", NewPageError(Err500Code, Err500Title, Err500Message)
	}

	// #nosec G203 -- Content is from trusted markdown files from local filesystem or controlled git repo, not user input
	return template.HTML(htmlString), nil
}

func getHTMLContent(app *App, pageName, mdPath string) (string, error) {
	if app.ServerConfig.HTMLCompilationMode == constants.HTMLCompilationModeStatic {
		htmlContentBytes, err := os.ReadFile(filepath.Clean(
			filepath.Join(
				app.ServerConfig.GeneratedPath,
				constants.HTMLFilesPath,
				pageName+".html",
			)),
		)
		if err != nil {
			app.Logger.Error("Error reading html content: %v", err)
			return "", NewPageError(Err500Code, Err500Title, Err500Message)
		}
		return string(htmlContentBytes), nil
	}

	compiledHTML, err := htmlcompiler.CompileHTMLFile(mdPath, app.SiteConfig)
	if err != nil {
		app.Logger.Error("Error compiling markdown: %v", err)
		return "", NewPageError(Err500Code, Err500Title, Err500Message)
	}
	return compiledHTML, nil
}
