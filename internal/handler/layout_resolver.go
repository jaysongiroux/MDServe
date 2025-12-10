package handler

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"

	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

func determineLayout(app *App, pageName string) (string, string) {
	layoutFile := defaultLayoutFile
	layoutFilter := ""

	for _, layout := range app.SiteConfig.Site.Layouts {
		match, err := regexp.MatchString(layout.Page, pageName)
		if err != nil {
			app.Logger.Error("Error matching layout page: %v", err)
			break
		}

		if match {
			return layout.Layout, layout.Filter
		}
	}

	return layoutFile, layoutFilter
}

func renderCustomLayout(
	app *App,
	w http.ResponseWriter,
	layoutFile, pageName, mdPath string,
	data *TemplateData,
) error {
	customLayoutName := layoutFile + ".html"
	app.Logger.Info("Using custom layout: %s", customLayoutName)

	// Handle blog article layout specifics
	if customLayoutName == blogArticleLayoutName {
		if err := prepareBlogArticleData(app, pageName, mdPath, data); err != nil {
			return err
		}
	}

	// Render the custom layout
	var customLayoutBuf strings.Builder
	if err := app.Templates.ExecuteTemplate(&customLayoutBuf, customLayoutName, data); err != nil {
		app.Logger.Error("Custom layout execution error: %v", err)
		return NewPageError(Err500Code, Err500Title, Err500Message)
	}

	// Nest within default layout
	// #nosec G203 -- Content is from Go template execution with trusted template files, not user input
	data.Content = template.HTML(customLayoutBuf.String())
	if err := app.Templates.ExecuteTemplate(w, defaultLayoutFile, data); err != nil {
		app.Logger.Error("Template execution error: %v", err)
		return err
	}

	data.PageName = &pageName

	return nil
}

func prepareBlogArticleData(app *App, pageName string, mdPath string, data *TemplateData) error {
	app.Logger.Info("Fetching headers for the blog article layout")

	htmlContent, err := getHTMLContent(app, pageName, mdPath)
	if err != nil {
		return err
	}

	headers, err := htmlcompiler.GetHeaders(htmlContent)
	if err != nil {
		app.Logger.Error("Error getting headers: %v", err)
		return NewPageError(Err500Code, Err500Title, Err500Message)
	}

	data.Headers = headers

	// get first h1 header
	firstHeader, err := htmlcompiler.GetFirstHeader("h1", htmlContent)
	if err != nil {
		app.Logger.Error("Error getting first header: %v", err)
		return NewPageError(Err500Code, Err500Title, Err500Message)
	}

	if firstHeader == "" {
		firstHeader = pageName
	}

	data.PageName = &firstHeader

	return nil
}
