package config

import (
	"fmt"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"go.yaml.in/yaml/v3"
)

type ServerConfig struct {
	Port                                int                           `yaml:"port"`
	Host                                string                        `yaml:"host"`
	HTMLCompilationMode                 constants.HTMLCompilationMode `yaml:"html_compilation_mode"`
	ContentPath                         string                        `yaml:"content_path"`
	AssetsPath                          string                        `yaml:"assets_path"`
	UserStaticPath                      string                        `yaml:"user_static_path"`
	TemplatesPath                       string                        `yaml:"templates_path"`
	GeneratedPath                       string                        `yaml:"generated_path"`
	LogLevel                            logger.LogLevel               `yaml:"log_level"`
	OptimizeImages                      bool                          `yaml:"optimize_images"`
	OptimizeImagesQuality               int                           `yaml:"optimize_images_quality"`
	Demo                                bool                          `yaml:"demo"`
	GitRemoteContentURL                 string                        `yaml:"git_remote_content_path"`
	GitRemoteContentDirectory           string                        `yaml:"git_remote_content_directory"`
	GitRemoteContentBranch              string                        `yaml:"git_remote_content_branch"`
	GitRemoteContentAssetsDirectory     string                        `yaml:"git_remote_content_assets_directory"`
	GitRemoteContentUserStaticDirectory string                        `yaml:"git_remote_content_user_static_directory"`
}

func LoadServerConfig() (*ServerConfig, error) {
	configContent, err := GetConfigContent(ConfigTypeServer)
	if err != nil {
		return nil, fmt.Errorf("failed to get config content: %w", err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server config: %w", err)
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate checks if the server configuration is valid
func (c *ServerConfig) Validate() error {
	// Check for mutually exclusive options
	if err := c.validateMutuallyExclusiveOptions(); err != nil {
		return err
	}

	// Validate git remote content configuration if enabled
	if c.GitRemoteContentURL != "" {
		if err := c.validateGitRemoteFields(); err != nil {
			return err
		}
	}

	return nil
}

// validateMutuallyExclusiveOptions ensures demo and git remote aren't both enabled
func (c *ServerConfig) validateMutuallyExclusiveOptions() error {
	if c.Demo && c.GitRemoteContentURL != "" {
		err := fmt.Errorf("demo mode and git remote content URL cannot be enabled at the same time")
		logger.Error(err.Error())
		return err
	}

	if c.Demo {
		logger.Warn("Demo mode is enabled. This will copy the README to the content folder as index.md")
	}

	return nil
}

// validateGitRemoteFields ensures at least one directory is configured and branch is always required
func (c *ServerConfig) validateGitRemoteFields() error {
	// Branch is always required
	if c.GitRemoteContentBranch == "" {
		err := fmt.Errorf("git_remote_content_branch is required when git remote content URL is provided")
		logger.Error(err.Error())
		return err
	}

	// At least one directory must be configured
	hasAtLeastOneDirectory := c.GitRemoteContentDirectory != "" ||
		c.GitRemoteContentAssetsDirectory != "" ||
		c.GitRemoteContentUserStaticDirectory != ""

	if !hasAtLeastOneDirectory {
		err := fmt.Errorf("at least one directory must be configured (git_remote_content_directory, git_remote_content_assets_directory, or git_remote_content_user_static_directory) when git remote content URL is provided")
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (c *ServerConfig) HasGitRemoteContentDirectory() bool {
	return c.GitRemoteContentDirectory != ""
}

func (c *ServerConfig) HasGitRemoteAssetsDirectory() bool {
	return c.GitRemoteContentAssetsDirectory != ""
}

func (c *ServerConfig) HasGitRemoteUserStaticDirectory() bool {
	return c.GitRemoteContentUserStaticDirectory != ""
}
