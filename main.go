package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/assets"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/demo"
	"github.com/jaysongiroux/mdserve/internal/files"
	"github.com/jaysongiroux/mdserve/internal/handler"
	htmlcompiler "github.com/jaysongiroux/mdserve/internal/html_compiler"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	appLogger := logger.New("Initial Setup", logger.DebugLevel)
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		appLogger.Error("No .env file found or unable to load it, using system environment variables")
	}

	app := &handler.App{
		ServerConfig: nil,
		SiteConfig:   nil,
		Logger:       nil,
		Templates:    nil,
		Handler:      handler.HandlePage,
	}

	serverConfig, err := config.LoadServerConfig()
	if err != nil {
		appLogger.Error("Failed to load server config: %v", err)
		panic(err)
	}
	app.ServerConfig = serverConfig

	siteConfig, err := config.LoadSiteConfig()
	if err != nil {
		appLogger.Error("Failed to load site config: %v", err)
		panic(err)
	}
	app.SiteConfig = siteConfig

	err = logger.Init(app.ServerConfig.LogLevel)
	if err != nil {
		appLogger.Error("Failed to initialize logger: %v", err)
		panic(err)
	}
	defer logger.Sync()

	appLogger = logger.New("Main", app.ServerConfig.LogLevel)
	app.Logger = appLogger

	if app.ServerConfig.Demo {
		err = demo.HandleDemoEnabled(app)
		if err != nil {
			appLogger.Error("Failed to handle demo mode: %v", err)
			return
		}
	}

	generatedPath := app.ServerConfig.GeneratedPath

	// Delete the generated path
	err = files.DeleteDirectory(generatedPath)
	if err != nil {
		appLogger.Error("Failed to delete generated path: %v", err)
		return
	}

	// if HTML Compilation mode is static, compile the HTML files
	if app.ServerConfig.HTMLCompilationMode == constants.HTMLCompilationModeStatic {
		appLogger.Info("Compiling static HTML files")
		mdFiles, err := htmlcompiler.GetMDFiles(app.ServerConfig.ContentPath)
		if err != nil {
			appLogger.Error("Failed to get MD files: %v", err)
			return
		}

		for _, mdFile := range mdFiles {
			htmlFile, err := htmlcompiler.CompileHTMLFile(mdFile, app.SiteConfig)
			appLogger.Debug("Compiling HTML file: %s", mdFile)
			if err != nil {
				appLogger.Error("Failed to compile HTML file: %v", err)
				return
			}

			// Calculate relative path from content directory to preserve folder structure
			relPath, err := filepath.Rel(app.ServerConfig.ContentPath, mdFile)
			if err != nil {
				appLogger.Error("Failed to get relative path for %s: %v", mdFile, err)
				return
			}

			// save the HTML file to the compiled HTML path
			// write files to the content path /.html/ preserving directory structure
			savePath := filepath.Join(generatedPath)
			appLogger.Debug("Writing HTML file: %s to %s", relPath, savePath)

			err = htmlcompiler.WriteHTMLFile(savePath, relPath, htmlFile)
			if err != nil {
				appLogger.Error("Failed to write HTML file: %v", err)
				return
			}
		}

		appLogger.Info("Static HTML files compiled successfully")
	}

	// handle asset optimization
	logger.Info("Optimizing assets")
	assetsGeneratedPath := filepath.Join(generatedPath, constants.GeneratedAssetsPath)
	logger.Info("Moving assets to generated path: %s", assetsGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.AssetsPath, assetsGeneratedPath)
	if err != nil {
		appLogger.Error("Failed to move assets: %v", err)
		return
	}

	// get all the assets that have been moved to the generated assets path
	allAssets, err := assets.GetAssets(assetsGeneratedPath)
	if err != nil {
		appLogger.Error("Failed to get all assets: %v", err)
		return
	}

	// find all assets that can be optimized
	optimizableAssets, err := assets.GetOptimizableAssets(allAssets)
	if err != nil {
		appLogger.Error("Failed to get optimizable assets: %v", err)
		return
	}

	// optimize all assets that can be optimized
	for _, asset := range optimizableAssets {
		err = assets.ConvertToWebP(asset, app.ServerConfig.OptimizeImages, app.ServerConfig.OptimizeImagesQuality)
		if err != nil {
			appLogger.Error("Failed to convert asset to WebP: %v", err)
			return
		}
		err = assets.DeleteAsset(asset)
		if err != nil {
			appLogger.Error("Failed to delete asset: %v", err)
			return
		}
	}
	logger.Info("Assets optimized successfully")

	// move all user-static assets to the generated path
	userStaticGeneratedPath := filepath.Join(generatedPath, constants.UserStaticPath)
	logger.Info("Moving user-static assets to generated path: %s", userStaticGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.UserStaticPath, userStaticGeneratedPath)
	if err != nil {
		appLogger.Error("Failed to move user-static assets: %v", err)
		return
	}
	logger.Info("User-static assets moved successfully")

	templatesGeneratedPath := filepath.Join(generatedPath, constants.TemplatesPath)
	logger.Info("Moving templates to generated path: %s", templatesGeneratedPath)
	err = assets.MoveAssets(app.ServerConfig.TemplatesPath, templatesGeneratedPath)
	if err != nil {
		appLogger.Error("Failed to move templates: %v", err)
		return
	}
	logger.Info("Templates moved successfully")

	logger.Info("Generated path: %s", generatedPath)
	siteMapPath := filepath.Join(generatedPath, constants.SiteMapPath)
	logger.Info("Site map path: %s", siteMapPath)

	// generate the site map
	siteMap, err := htmlcompiler.GenerateSiteMap(app.ServerConfig.ContentPath, app.SiteConfig)
	if err != nil {
		appLogger.Error("Failed to generate site map: %v", err)
		return
	}
	err = htmlcompiler.SaveSiteMap(siteMap, siteMapPath)
	if err != nil {
		appLogger.Error("Failed to save site map: %v", err)
		return
	}
	logger.Info("Site map saved successfully to %s", siteMapPath)

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
	app.Templates, err = app.Templates.ParseGlob(templatesGeneratedPath + "/*.html")
	if err != nil {
		app.Logger.Fatal("Failed to parse templates: %v", err)
	}

	// load templates from layout_templates subdirectory
	layoutTemplatesPath := filepath.Join(templatesGeneratedPath, "layout_templates")
	if _, err := os.Stat(layoutTemplatesPath); err == nil {
		app.Templates, err = app.Templates.ParseGlob(layoutTemplatesPath + "/*.html")
		if err != nil {
			app.Logger.Fatal("Failed to parse layout templates: %v", err)
		}
	}

	mux := http.NewServeMux()

	// Serve Optimized System Assets (Mapped to /assets/)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsGeneratedPath))))

	// Serve User Static Assets (Mapped to /user-static/)
	// Assuming this config exists in ServerConfig
	userStaticPath := "user-static"
	mux.Handle("GET /user-static/", http.StripPrefix("/user-static/", http.FileServer(http.Dir(userStaticPath))))

	// Main Page Handler
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		app.Handler(app, w, r)
	})

	// --- 5. Start Server ---
	port := app.ServerConfig.Port
	app.Logger.Info("Starting MDServe on port %d in %s mode", port, app.ServerConfig.HTMLCompilationMode)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		app.Logger.Fatal("Server failed: %v", err)
	}
}
