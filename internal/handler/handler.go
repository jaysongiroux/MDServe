package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/constants"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

var (
	Err404Code    = "404"
	Err500Code    = "500"
	Err404Title   = "Page Not Found"
	Err404Message = "The page you're looking for doesn't exist. It might have been moved, deleted, or you entered the wrong URL."
	Err500Title   = "Internal Server Error"
	Err500Message = "An unexpected error occurred while processing your request."
)

const (
	defaultLayoutFile     = "layout.html"
	blogArticleLayoutName = "blog_article_layout.html"
)

func HandlePage(app *App, w http.ResponseWriter, r *http.Request) {
	pageName := getPageName(r.URL.Path)

	// Initialize template data
	data := newTemplateData(app)

	// Load page content
	mdPath := filepath.Join(app.ServerConfig.ContentPath, pageName+".md")
	contentHTML, err := loadPageContent(app, pageName, mdPath)
	if err != nil {
		handleError(app, w, err, &data)
		return
	}
	data.Content = contentHTML

	// Load sitemap metadata
	sitemapPath := filepath.Join(app.ServerConfig.GeneratedPath, constants.SiteMapPath)
	sitemapEntity, err := htmlcompiler.GetSitemapEntityByPath(pageName, sitemapPath)
	if err != nil {
		handleError(app, w, err, &data)
		return
	}

	data.CreationDate = htmlcompiler.GetCreationDate(*sitemapEntity)
	if sitemapEntity.Metadata != nil {
		data.Metadata = sitemapEntity.Metadata
	} else {
		data.Metadata = nil
	}

	fmt.Println("metadata", data.Metadata)

	// Determine layout
	layoutFile, layoutFilter := determineLayout(app, pageName)

	// Apply layout filter if specified
	if layoutFilter != "" {
		if err := applyLayoutFilter(app, layoutFilter, sitemapPath, &data); err != nil {
			handleError(app, w, err, &data)
			return
		}
	}

	// Handle custom layouts
	if layoutFile != defaultLayoutFile {
		if err := renderCustomLayout(app, w, layoutFile, pageName, mdPath, &data); err != nil {
			handleError(app, w, err, &data)
			return
		}
	} else {
		app.Logger.Info("Using default layout: %s", defaultLayoutFile)
		if err := app.Templates.ExecuteTemplate(w, defaultLayoutFile, data); err != nil {
			app.Logger.Error("Template execution error: %v", err)
		}
	}

	w.Header().Set("Content-Type", "text/html")
}

func getPageName(path string) string {
	if path == "/" {
		return "index"
	}

	pageName := strings.TrimPrefix(path, "/")
	if strings.HasSuffix(pageName, "/") {
		return pageName + "index"
	}
	return pageName
}
