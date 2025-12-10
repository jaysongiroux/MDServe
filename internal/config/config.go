package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	// can be a http request url which returns a yaml file
	// or a git https url that has to end in .git
	ENV_VAR_MD_SERVER_CONFIG_PATH = "MD_SERVER_CONFIG_PATH"
	ENV_VAR_MD_SITE_CONFIG_PATH   = "MD_SITE_CONFIG_PATH"
	// empty defaults to default branch which is usually master or main
	ENV_VAR_MD_CONFIG_BRANCH = "MD_CONFIG_BRANCH"
	// directory in the remote path to fetch the config from
	// ex. config or config/nested_directory
	ENV_VAR_MD_CONFIG_LOCATION = "MD_CONFIG_LOCATION"
	// plain text username
	ENV_VAR_GIT_USERNAME = "GIT_USERNAME"
	// PAT is preferred over password
	ENV_VAR_GIT_PASSWORD = "GIT_PASSWORD"
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
	response, err := http.Get(filepath.Clean(configPath))
	if err != nil {
		return "", err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logger.Error("Failed to close response body: %v", err)
		}
	}()

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
	if strings.HasSuffix(configPath, ".git") {
		branch := os.Getenv(ENV_VAR_MD_CONFIG_BRANCH)
		configLocation := os.Getenv(ENV_VAR_MD_CONFIG_LOCATION)

		logger.Debug("Fetching remote config from: %s", configPath)
		configFileName := ""
		switch configType {
		case ConfigTypeServer:
			configFileName = constants.ServerConfigName
		case ConfigTypeSite:
			configFileName = constants.SiteConfigName
		default:
			return "", fmt.Errorf("invalid config type: %s", configType)
		}

		configContent, err := GetFileFromRepo(configPath, &branch, configLocation, configFileName)
		if err != nil {
			return "", fmt.Errorf("failed to fetch remote config: %w", err)
		}
		return configContent, nil
	} else if strings.HasPrefix(configPath, "http") {
		configContent, err := getRemoteConfigContent(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to fetch remote config: %w", err)
		}
		return configContent, nil
	}

	// default to local file path
	configContent, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return "", fmt.Errorf("failed to read local config: %w", err)
	}
	logger.Debug("Successfully read local config: %s", configPath)
	return string(configContent), nil
}
