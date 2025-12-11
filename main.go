package main

// Abstraction layer for zap logger
// Provides a simple interface for logging messages
// with different levels of verbosity
// and a cron logger adapter for cron jobs

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jaysongiroux/mdserve/internal/assets"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/demo"
	"github.com/jaysongiroux/mdserve/internal/files"
	"github.com/jaysongiroux/mdserve/internal/git"
	"github.com/jaysongiroux/mdserve/internal/handler"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func prelimSetup(callerName string) (*handler.App, error) {
	appLogger := logger.New("Initial Setup", logger.DebugLevel)

	app := &handler.App{
		ServerConfig:            nil,
		SiteConfig:              nil,
		Logger:                  nil,
		Templates:               nil,
		Handler:                 handler.HandlePage,
		TemplatesGeneratedPath:  "",
		AssetsGeneratedPath:     "",
		UserStaticGeneratedPath: "",
	}

	serverConfig, err := config.LoadServerConfig()
	if err != nil {
		appLogger.Fatal("Failed to load server config: %v", err)
	}
	app.ServerConfig = serverConfig

	siteConfig, err := config.LoadSiteConfig()
	if err != nil {
		appLogger.Fatal("Failed to load site config: %v", err)
	}
	app.SiteConfig = siteConfig

	err = logger.Init(app.ServerConfig.LogLevel)
	if err != nil {
		appLogger.Fatal("Failed to initialize logger: %v", err)
	}

	defer func() {
		err = logger.Sync()
		if err != nil {
			// Ignore errors from syncing stdout/stderr in non-TTY environments (Docker, pipes, etc.)
			if errors.Is(err, syscall.ENOTTY) || errors.Is(err, syscall.EINVAL) ||
				errors.Is(err, os.ErrInvalid) {
				return
			}
			appLogger.Fatal("Failed to sync logger: %v", err)
		}
	}()

	appLogger = logger.New(callerName, app.ServerConfig.LogLevel)
	app.Logger = appLogger

	if app.ServerConfig.Demo {
		err = demo.HandleDemoEnabled(app)
		if err != nil {
			appLogger.Fatal("Failed to handle demo mode: %v", err)
		}
	}

	if app.ServerConfig.GitRemoteContentURL != "" {
		err = git.HandleGitRemoteContent(app.ServerConfig)
		if err != nil {
			appLogger.Fatal("Failed to handle git remote content: %v", err)
		}
	}

	// Delete the generated path
	err = files.DeleteDirectoryContents(app.ServerConfig.GeneratedPath)
	if err != nil {
		appLogger.Fatal("Failed to delete generated path: %v", err)
	}

	// if HTML Compilation mode is static, compile the HTML files
	if app.ServerConfig.HTMLCompilationMode == constants.HTMLCompilationModeStatic {
		appLogger.Info("Compiling static HTML files")
		mdFiles, err := htmlcompiler.GetMDFiles(app.ServerConfig.ContentPath)
		if err != nil {
			appLogger.Fatal("Failed to get MD files: %v", err)
		}

		for _, mdFile := range mdFiles {
			htmlFile, err := htmlcompiler.CompileHTMLFile(mdFile, app.SiteConfig)
			appLogger.Debug("Compiling HTML file: %s", mdFile)
			if err != nil {
				appLogger.Fatal("Failed to compile HTML file: %v", err)
			}

			// Calculate relative path from content directory to preserve folder structure
			relPath, err := filepath.Rel(app.ServerConfig.ContentPath, mdFile)
			if err != nil {
				appLogger.Fatal("Failed to get relative path for %s: %v", mdFile, err)
			}

			// save the HTML file to the compiled HTML path
			// write files to the content path /.html/ preserving directory structure
			savePath := filepath.Join(app.ServerConfig.GeneratedPath)
			appLogger.Debug("Writing HTML file: %s to %s", relPath, savePath)

			err = htmlcompiler.WriteHTMLFile(savePath, relPath, htmlFile)
			if err != nil {
				appLogger.Fatal("Failed to write HTML file: %v", err)
			}
		}

		appLogger.Info("Static HTML files compiled successfully")
	}

	// handle asset optimization
	logger.Info("Optimizing assets")
	app.AssetsGeneratedPath = filepath.Join(
		app.ServerConfig.GeneratedPath,
		constants.GeneratedAssetsPath,
	)
	logger.Info("Moving assets to generated path: %s", app.AssetsGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.AssetsPath, app.AssetsGeneratedPath)
	if err != nil {
		appLogger.Fatal("Failed to move assets: %v", err)
	}

	// get all the assets that have been moved to the generated assets path
	allAssets, err := files.GetAllFilesInDirectory(app.AssetsGeneratedPath)
	if err != nil {
		appLogger.Fatal("Failed to get all assets: %v", err)
	}

	// find all assets that can be optimized
	optimizableAssets, err := assets.GetOptimizableAssets(allAssets)
	if err != nil {
		appLogger.Fatal("Failed to get optimizable assets: %v", err)
	}

	siteManifestIconPaths, err := assets.GetIconPathsFromSiteWebmanifest(
		filepath.Join(app.ServerConfig.AssetsPath, "site.webmanifest"),
	)
	if err != nil {
		appLogger.Fatal("Failed to get site manifest icon paths: %v", err)
	}

	logger.Debug("Site manifest icon paths: %v", siteManifestIconPaths)

	// optimize all assets that can be optimized
	for _, asset := range optimizableAssets {
		// filter out the assets related to the site.webmanifest
		logger.Info("Checking asset: %s", asset)
		if assets.ShouldSkipAsset(asset, siteManifestIconPaths) {
			logger.Info(
				"Skipping asset: %s because it is a site manifest icon, favicon, SVG, or a format that is not supported for optimization",
				asset,
			)
			continue
		}
		err = assets.ConvertToWebP(
			asset,
			app.ServerConfig.OptimizeImages,
			app.ServerConfig.OptimizeImagesQuality,
		)
		if err != nil {
			appLogger.Fatal("Failed to convert asset to WebP: %v", err)
		}
		err = assets.DeleteAsset(asset)
		if err != nil {
			appLogger.Fatal("Failed to delete asset: %v", err)
		}
	}

	logger.Info("Assets optimized successfully")

	// move all user-static assets to the generated path
	app.UserStaticGeneratedPath = filepath.Join(
		app.ServerConfig.GeneratedPath,
		constants.UserStaticPath,
	)
	logger.Info("Moving user-static assets to generated path: %s", app.UserStaticGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.UserStaticPath, app.UserStaticGeneratedPath)
	if err != nil {
		appLogger.Fatal("Failed to move user-static assets: %v", err)
	}
	logger.Info("User-static assets moved successfully")

	app.TemplatesGeneratedPath = filepath.Join(
		app.ServerConfig.GeneratedPath,
		constants.TemplatesPath,
	)
	logger.Info("Moving templates to generated path: %s", app.TemplatesGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.TemplatesPath, app.TemplatesGeneratedPath)
	if err != nil {
		appLogger.Fatal("Failed to move templates: %v", err)
	}
	logger.Info("Templates moved successfully")

	logger.Info("Generated path: %s", app.ServerConfig.GeneratedPath)
	siteMapPath := filepath.Join(app.ServerConfig.GeneratedPath, constants.SiteMapPath)
	logger.Info("Site map path: %s", siteMapPath)

	// generate the site map
	siteMap, err := htmlcompiler.GenerateSiteMap(app.ServerConfig.ContentPath, app.SiteConfig)
	if err != nil {
		appLogger.Fatal("Failed to generate site map: %v", err)
	}
	err = htmlcompiler.SaveSiteMap(siteMap, siteMapPath)
	if err != nil {
		appLogger.Fatal("Failed to save site map: %v", err)
	}
	logger.Info("Site map saved successfully to %s", siteMapPath)

	return app, nil
}

func main() {
	appLogger := logger.New("Initial Setup", logger.DebugLevel)
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		appLogger.Warn(
			"No .env file found or unable to load it, using system environment variables",
		)
	}

	app, err := prelimSetup("Main")
	if err != nil {
		appLogger.Fatal("Failed to perform prelim setup: %v", err)
	}

	app.Logger.Info("Loading HTML templates...")

	app.Templates = template.New("").Funcs(template.FuncMap{
		"table_of_contents_href": func(text string) string {
			// lower and replace all non-alphanumeric characters (except spaces) with ""
			re := regexp.MustCompile(`[^a-zA-Z0-9 ]`)
			cleanedText := re.ReplaceAllString(strings.TrimSpace(text), "")
			// replace spaces with hyphens
			cleanedText = strings.ReplaceAll(cleanedText, " ", "-")
			return strings.ToLower(cleanedText)
		},
		"is_last": func(index int, length int) bool {
			return index == int(length)-1
		},
		"array_to_string": func(array []string) string {
			return strings.Join(array, ", ")
		},
	})
	app.Templates, err = app.Templates.ParseGlob(app.TemplatesGeneratedPath + "/*.html")
	if err != nil {
		app.Logger.Fatal("Failed to parse templates: %v", err)
	}

	// load templates from layout_templates subdirectory
	layoutTemplatesPath := filepath.Join(app.TemplatesGeneratedPath, "layout_templates")
	if _, err := os.Stat(layoutTemplatesPath); err == nil {
		app.Templates, err = app.Templates.ParseGlob(layoutTemplatesPath + "/*.html")
		if err != nil {
			app.Logger.Fatal("Failed to parse layout templates: %v", err)
		}
	}

	if app.ServerConfig.GenerationCronEnabled {
		app.Logger.Info(
			"Generation cron is enabled. Starting cron scheduler using interval %s...",
			app.ServerConfig.GenerationCronInterval,
		)
		// Initialize and start cron scheduler
		cronLogger := app.Logger.ToCronLogger(app.ServerConfig.LogLevel == logger.DebugLevel)
		c := cron.New(cron.WithChain(
			cron.Recover(cronLogger),
		))

		// Add hourly job
		_, err = c.AddFunc(app.ServerConfig.GenerationCronInterval, func() {
			app.Logger.Info("Running hourly cron job...")
			_, err = prelimSetup("Generation Cron")
			if err != nil {
				app.Logger.Fatal("Failed to perform prelim setup: %v", err)
			}
		})
		if err != nil {
			app.Logger.Fatal("Failed to schedule cron job: %v", err)
		}

		c.Start()
		app.Logger.Info("Cron scheduler started - hourly job registered")

		// Ensure cron stops when main exits
		defer func() {
			err := c.Stop()
			if err != nil {
				app.Logger.Fatal("Failed to stop cron scheduler: %v", err)
			}
		}()
	}

	mux := http.NewServeMux()

	// Serve Optimized System Assets (Mapped to /assets/)
	mux.Handle(
		"GET /assets/",
		http.StripPrefix("/assets/", http.FileServer(http.Dir(app.AssetsGeneratedPath))),
	)

	// Serve User Static Assets (Mapped to /user-static/)
	// Assuming this config exists in ServerConfig
	userStaticPath := "user-static"
	mux.Handle(
		"GET /user-static/",
		http.StripPrefix("/user-static/", http.FileServer(http.Dir(userStaticPath))),
	)

	mux.HandleFunc("GET /sitemap.xml", handler.HandleSitemap(app))

	// Main Page Handler
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		app.Handler(app, w, r)
	})

	// --- 5. Start Server ---
	port := app.ServerConfig.Port
	app.Logger.Info(
		"Starting MDServe on port %d in %s mode",
		port,
		app.ServerConfig.HTMLCompilationMode,
	)

	// Wrap mux with middleware
	// Order: CORS -> Cache -> Mux
	srvHandler := handler.AddCORSHeaders(handler.AddCacheHeaders(app, mux))

	// add timeout to the server
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      srvHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		app.Logger.Fatal("Server failed: %v", err)
	}
}
