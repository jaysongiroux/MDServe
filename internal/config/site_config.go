package config

import (
	"os"

	"github.com/jaysongiroux/mdserve/internal/constants"
	"go.yaml.in/yaml/v3"
)

// define config structs
type SiteConfig struct {
	Navbar []NavbarItem `yaml:"navbar"`
	Footer Footer       `yaml:"footer"`
	Site   Site         `yaml:"site"`
}

type Footer struct {
	Copyright string `yaml:"copyright"`
	Links     []Link `yaml:"links"`
}

type Site struct {
	Name                      string        `yaml:"name"`
	PoweredBy                 string        `yaml:"powered_by"`
	PoweredByURL              string        `yaml:"powered_by_url"`
	Theme                     Theme         `yaml:"theme"`
	Layouts                   []Layout      `yaml:"layouts"`
	PageSize                  int           `yaml:"page_size"`
	SortDirection             SortDirection `yaml:"sort_direction"`
	Keywords                  []string      `yaml:"keywords"`
	Author                    string        `yaml:"author"`
	Description               string        `yaml:"description"`
	AllowSearchEngineIndexing bool          `yaml:"allow_search_engine_indexing"`
}

type Layout struct {
	Page   string `yaml:"page"`
	Filter string `yaml:"filter"`
	Layout string `yaml:"layout"`
}

type NavbarItem struct {
	Label    string `yaml:"label"`
	URL      string `yaml:"url"`
	Dropdown []Link `yaml:"dropdown"`
}

type Link struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

type Theme struct {
	Code Code `yaml:"code"`
}

type Code struct {
	Theme       string `yaml:"theme"`
	LineNumbers bool   `yaml:"line_numbers"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

func LoadSiteConfig() (*SiteConfig, error) {
	yamlFile, err := os.ReadFile(constants.SiteConfigPath)
	if err != nil {
		return nil, err
	}

	var siteConfig SiteConfig
	err = yaml.Unmarshal(yamlFile, &siteConfig)
	if err != nil {
		return nil, err
	}

	return &siteConfig, nil
}
