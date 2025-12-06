package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

type ConfigType string

const (
	ConfigTypeServer ConfigType = "server"
	ConfigTypeSite   ConfigType = "site"
)

const (
	ENV_VAR_MD_SERVER_CONFIG_PATH = "MD_SERVER_CONFIG_PATH"
	ENV_VAR_MD_SITE_CONFIG_PATH   = "MD_SITE_CONFIG_PATH"
)

func getConfigPath(defaultPath string, envVariable string) string {
	envValue := os.Getenv(envVariable)
	if envValue != "" {
		logger.Debug("Using config path from environment variable %s: %s", envVariable, envValue)
		return envValue
	}
	logger.Debug("Using default config path: %s", defaultPath)
	return defaultPath
}

func getRemoteConfigContent(configPath string) (string, error) {
	logger.Debug("Fetching remote config from: %s", configPath)
	response, err := http.Get(configPath)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch config: %d", response.StatusCode)
	}

	logger.Debug("Successfully fetched remote config (%d bytes)", len(body))
	return string(body), nil
}

func GetConfigContent(configType ConfigType) (string, error) {
	var configPath string
	switch configType {
	case ConfigTypeServer:
		configPath = getConfigPath(constants.ServerConfigPath, ENV_VAR_MD_SERVER_CONFIG_PATH)
	case ConfigTypeSite:
		configPath = getConfigPath(constants.SiteConfigPath, ENV_VAR_MD_SITE_CONFIG_PATH)
	default:
		return "", fmt.Errorf("invalid config type: %s", configType)
	}

	logger.Debug("Loading %s config from: %s", configType, configPath)

	if strings.HasPrefix(configPath, "http") {
		configContent, err := getRemoteConfigContent(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to fetch remote config: %w", err)
		}
		return configContent, nil
	}

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read local config: %w", err)
	}
	return string(configContent), nil
}
