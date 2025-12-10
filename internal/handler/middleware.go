package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jaysongiroux/mdserve/internal/constants"
)

func AddCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AddCacheHeaders(app *App, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if it's a static asset
		isStatic := strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/user-static/") ||
			strings.HasSuffix(path, ".xml") ||
			strings.HasSuffix(path, ".txt")

		if isStatic {
			cacheDuration := time.Duration(app.ServerConfig.CacheStaticMaxAge) * time.Second
			w.Header().
				Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int(cacheDuration.Seconds())))
		} else {
			// Dynamic content (HTML)
			w.Header().Set("Content-Type", "text/html")

			var cacheDuration time.Duration
			if app.ServerConfig.HTMLCompilationMode == constants.HTMLCompilationModeLive {
				cacheDuration = time.Duration(app.ServerConfig.CacheHTMLMaxAge) * time.Second
			} else {
				// Even if static compilation, HTML files might change on redeploy, but we treat them as static-ish?
				// The original code had logic for this.
				// "cache_html_max_age" is usually for HTML.
				cacheDuration = time.Duration(app.ServerConfig.CacheHTMLMaxAge) * time.Second
			}

			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, s-maxage=%d", int(cacheDuration.Seconds()), int(cacheDuration.Seconds())))
			w.Header().Set("Vary", "Accept-Encoding")
		}

		next.ServeHTTP(w, r)
	})
}
