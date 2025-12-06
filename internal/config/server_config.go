package config

import (
	"os"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"go.yaml.in/yaml/v3"
)

type ServerConfig struct {
	Port                  int                           `yaml:"port"`
	Host                  string                        `yaml:"host"`
	HTMLCompilationMode   constants.HTMLCompilationMode `yaml:"html_compilation_mode"`
	ContentPath           string                        `yaml:"content_path"`
	AssetsPath            string                        `yaml:"assets_path"`
	UserStaticPath        string                        `yaml:"user_static_path"`
	TemplatesPath         string                        `yaml:"templates_path"`
	GeneratedPath         string                        `yaml:"generated_path"`
	LogLevel              logger.LogLevel               `yaml:"log_level"`
	OptimizeImages        bool                          `yaml:"optimize_images"`
	OptimizeImagesQuality int                           `yaml:"optimize_images_quality"`
	Demo                  bool                          `yaml:"demo"`
}

func LoadServerConfig() (*ServerConfig, error) {
	yamlFile, err := os.ReadFile(constants.ServerConfigPath)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
