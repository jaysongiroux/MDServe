package handler

import (
	"html/template"
	"net/http"

	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

type App struct {
	ServerConfig *config.ServerConfig
	SiteConfig   *config.SiteConfig
	Logger       *logger.Logger
	Templates    *template.Template
	Handler      func(app *App, w http.ResponseWriter, r *http.Request)
}
