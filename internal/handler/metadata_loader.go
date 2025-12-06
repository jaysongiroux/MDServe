package handler

import (
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

func applyLayoutFilter(app *App, layoutFilter, sitemapPath string, data *TemplateData) error {
	siteMap, err := htmlcompiler.LoadSiteMap(sitemapPath)
	if err != nil {
		app.Logger.Error("Error loading site map: %v", err)
		return NewPageError(Err500Code, Err500Title, Err500Message)
	}

	siteMap, err = htmlcompiler.FilterSiteMap(siteMap, layoutFilter)
	if err != nil {
		app.Logger.Error("Error filtering site map: %v", err)
		return NewPageError(Err500Code, Err500Title, Err500Message)
	}

	data.PageList = siteMap
	return nil
}
