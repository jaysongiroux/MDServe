package handler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaysongiroux/mdserve/internal/constants"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
)

type SitemapXML struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URLXML `xml:"url"`
}

type URLXML struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func HandleSitemap(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sitemapPath := filepath.Join(app.ServerConfig.GeneratedPath, constants.SiteMapPath)
		entries, err := htmlcompiler.LoadSiteMap(sitemapPath)
		if err != nil {
			app.Logger.Error("Failed to load sitemap: %v", err)
			http.Error(w, "Failed to load sitemap", http.StatusInternalServerError)
			return
		}

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		baseURL := fmt.Sprintf("%s://%s", scheme, host)

		// Check if BaseURL is configured in SiteConfig (if we added it, but we didn't yet)
		// For now, rely on request host which is common for simple servers.

		var urls []URLXML
		for _, entry := range *entries {
			// Ensure path starts with /
			path := entry.Path
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			// Construct full URL
			loc := baseURL + path
            
            // Handle root path empty check or similar if needed, but usually path is like "about" or "blog/..."
            // If path is "index", it might be root. Logic in htmlcompiler usually handles index.
            // Let's assume entry.Path is correct relative path.

			lastMod := entry.LastModifiedDate.Format(time.RFC3339)
            if entry.LastModifiedDate.IsZero() {
                lastMod = entry.CreationDate.Format(time.RFC3339)
            }
            if lastMod == "0001-01-01T00:00:00Z" {
                 // Fallback to now if totally missing? Or skip?
                 // Let's use current time or omit? Required field usually.
                 lastMod = time.Now().Format(time.RFC3339)
            }

			urls = append(urls, URLXML{
				Loc:     loc,
				LastMod: lastMod,
				// ChangeFreq and Priority could be inferred or hardcoded
				ChangeFreq: "weekly", 
				Priority:   "0.5",
			})
		}

		sitemap := SitemapXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		w.Header().Set("Content-Type", "application/xml")
		// Cache control is handled by middleware, but we can ensure it here too or let middleware handle it.
        // Middleware checks extension .xml so it should be fine.
        
		encoder := xml.NewEncoder(w)
		encoder.Indent("", "  ")
		if err := encoder.Encode(sitemap); err != nil {
			app.Logger.Error("Failed to encode sitemap: %v", err)
			return
		}
	}
}

