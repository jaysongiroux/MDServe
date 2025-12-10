package htmlcompiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/files"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

func SortSiteMap(
	siteMap []SiteMapEntry,
	sortDirection config.SortDirection,
) (*[]SiteMapEntry, error) {
	if len(siteMap) == 0 {
		logger.Debug("Site map is empty, returning empty site map")
		return &siteMap, nil
	}

	if sortDirection != config.SortDirectionAsc && sortDirection != config.SortDirectionDesc {
		return nil, fmt.Errorf("invalid sort direction: %s", sortDirection)
	}

	sort.Slice(siteMap, func(i, j int) bool {
		var iDate, jDate time.Time

		// Get the creation dates for sorting
		iDate = GetCreationDate(siteMap[i])
		jDate = GetCreationDate(siteMap[j])

		// Handle zero dates - items with dates should come before items without dates
		if iDate.IsZero() && jDate.IsZero() {
			// Both zero, fall back to CreationDate field
			iDate = siteMap[i].CreationDate
			jDate = siteMap[j].CreationDate
		} else if iDate.IsZero() {
			// i has no date, j should come first
			return sortDirection == "asc"
		} else if jDate.IsZero() {
			// j has no date, i should come first
			return sortDirection == "desc"
		}

		// Both have dates, compare based on sort direction
		switch sortDirection {
		case "asc":
			return iDate.Before(jDate)
		case "desc":
			return iDate.After(jDate)
		default:
			return false
		}
	})

	return &siteMap, nil
}

func GenerateSiteMap(
	markdownFilePath string,
	siteConfig *config.SiteConfig,
) (*[]SiteMapEntry, error) {
	// crawl the markdown file path to get all mark down files
	mdFiles, err := GetMDFiles(markdownFilePath)
	if err != nil {
		return nil, err
	}

	var siteMap []SiteMapEntry

	for _, file := range mdFiles {
		// convert the markdown file to an HTML string
		htmlContent, err := CompileHTMLFile(file, siteConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to compile HTML file %s: %w", file, err)
		}

		firstHeader, err := GetFirstHeader("h1", htmlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to get first header for file %s: %w", file, err)
		}

		firstParagraph, err := GetFirstParagraph(htmlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to get first paragraph for file %s: %w", file, err)
		}

		markdownContent, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, fmt.Errorf("failed to read markdown content for file %s: %w", file, err)
		}
		metadata, err := GetMetadata(string(markdownContent))
		if err != nil {
			logger.Warn("failed to get metadata for file %s: %v", file, err)
			metadata = nil
		}

		var lastModifiedDate time.Time
		fileModifiedDate, err := files.GetFileModifiedDate(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get modified date for file %s: %w", file, err)
		}
		// handle if the modification date is not set in the metadata
		if metadata != nil {
			// check if LastModificationDate is set and not zero
			if metadata.LastModificationDate.IsZero() {
				lastModifiedDate = fileModifiedDate
			} else {
				lastModifiedDate = metadata.LastModificationDate
			}
		} else {
			lastModifiedDate = fileModifiedDate
		}

		var creationDate time.Time
		if metadata != nil {
			if metadata.CreationDate != (time.Time{}) {
				creationDate = metadata.CreationDate
			} else {
				creationDate = fileModifiedDate
			}
		} else {
			creationDate = fileModifiedDate
		}

		// remove md extention
		formattedPath := strings.TrimSuffix(file, ".md")
		// remove the content/ prefix
		formattedPath = strings.TrimPrefix(formattedPath, "content/")
		// replace spaces with _
		formattedPath = strings.ReplaceAll(formattedPath, " ", "_")

		// if path ends in /index remove the /index
		formattedPath = strings.TrimSuffix(formattedPath, "/index")

		siteMap = append(siteMap, SiteMapEntry{
			Path:             formattedPath,
			FirstHeader:      firstHeader,
			FirstParagraph:   firstParagraph,
			LastModifiedDate: lastModifiedDate,
			Metadata:         metadata,
			CreationDate:     creationDate,
		})
	}

	if siteMap != nil {
		logger.Debug("Sorting site map with sort direction: %s", siteConfig.Site.SortDirection)
		sortedSiteMap, err := SortSiteMap(siteMap, siteConfig.Site.SortDirection)
		if err != nil {
			logger.Error("Failed to sort site map: %v", err)
			return nil, fmt.Errorf("failed to sort site map: %w", err)
		}

		logger.Debug("Successfully sorted site map")

		return sortedSiteMap, nil
	}

	return &siteMap, nil
}

func SaveSiteMap(siteMap *[]SiteMapEntry, filePath string) error {
	jsonContent, err := json.Marshal(siteMap)
	if err != nil {
		return err
	}
	// with indentation
	err = os.WriteFile(filePath, jsonContent, 0600)
	if err != nil {
		return fmt.Errorf("failed to save site map to %s: %w", filePath, err)
	}
	return nil
}

func LoadSiteMap(filePath string) (*[]SiteMapEntry, error) {
	jsonContent, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	var siteMap []SiteMapEntry
	err = json.Unmarshal(jsonContent, &siteMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal site map from %s: %w", filePath, err)
	}

	return &siteMap, nil
}

func FilterSiteMap(siteMap *[]SiteMapEntry, regex string) (*[]SiteMapEntry, error) {
	regexObj, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}
	filteredSiteMap := make([]SiteMapEntry, 0)
	for _, article := range *siteMap {
		if regexObj.MatchString(article.Path) {
			filteredSiteMap = append(filteredSiteMap, article)
		}
	}

	return &filteredSiteMap, nil
}

// args
// path: the path of the page to get the sitemap entity for
// siteMapPath: the path to the site map file
func GetSitemapEntityByPath(path string, siteMapPath string) (*SiteMapEntry, error) {
	logger.Debug("Getting sitemap entity by path: %s", path)

	siteMap, err := LoadSiteMap(siteMapPath)
	if err != nil {
		logger.Error("Failed to load sitemap: %v", err)
		return nil, err
	}

	for _, page := range *siteMap {
		if page.Path == path {
			logger.Debug("Found page in sitemap: %s", path)
			return &page, nil
		}
	}

	logger.Warn("Page not found in sitemap: %s", path)
	return nil, errors.New("page not found in site map")
}
