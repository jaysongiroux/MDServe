package handler

import (
	"html/template"
	"time"

	"github.com/jaysongiroux/mdserve/internal/config"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

type TemplateData struct {
	PageName       *string
	Headers        []htmlcompiler.Header
	CreationDate   time.Time
	Site           config.Site
	Navbar         []config.NavbarItem
	Footer         config.Footer
	Content        template.HTML
	PoweredBy      string
	PoweredByURL   string
	UserStaticPath string
	AssetsPath     string
	ErrorCode      *string
	ErrorTitle     *string
	ErrorMessage   *string
	PageList       *[]htmlcompiler.SiteMapEntry
	Metadata       *htmlcompiler.Metadata
	SiteMapEntity  *htmlcompiler.SiteMapEntry
}

func newTemplateData(app *App) TemplateData {
	return TemplateData{
		Site:           app.SiteConfig.Site,
		Navbar:         app.SiteConfig.Navbar,
		Footer:         app.SiteConfig.Footer,
		PoweredBy:      app.SiteConfig.Site.PoweredBy,
		PoweredByURL:   app.SiteConfig.Site.PoweredByURL,
		UserStaticPath: "/" + app.ServerConfig.UserStaticPath,
		AssetsPath:     "/" + app.ServerConfig.AssetsPath,
		SiteMapEntity:  nil,
	}
}
