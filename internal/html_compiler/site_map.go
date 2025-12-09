package htmlcompiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/files"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

func SortSiteMap(siteMap []SiteMapEntry, sortDirection config.SortDirection) (*[]SiteMapEntry, error) {
	if len(siteMap) == 0 {
		return &siteMap, nil
	}

	sortFunc := func(i, j int) bool {
		var iDate, jDate time.Time

		// setting the dates for sorting
		iDate = GetCreationDate(siteMap[i])
		jDate = GetCreationDate(siteMap[j])

		// Sorting
		if !iDate.IsZero() && !jDate.IsZero() {
			return iDate.Before(jDate)
		}
		if !iDate.IsZero() && jDate.IsZero() {
			return true
		}
		if iDate.IsZero() && !jDate.IsZero() {
			return false
		}
		return siteMap[i].CreationDate.Before(siteMap[j].CreationDate)
	}

	if sortDirection == "asc" {
		sort.Slice(siteMap, sortFunc)
	} else if sortDirection == "desc" {
		sortedSiteMap := make([]SiteMapEntry, len(siteMap))
		copy(sortedSiteMap, siteMap)
		sort.Slice(sortedSiteMap, func(i, j int) bool {
			return sortFunc(j, i)
		})
		siteMap = sortedSiteMap
	} else {
		return nil, fmt.Errorf("invalid sort direction: %s", sortDirection)
	}

	return &siteMap, nil
}

func GenerateSiteMap(markdownFilePath string, siteConfig *config.SiteConfig) (*[]SiteMapEntry, error) {
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
			return nil, fmt.Errorf("failed to compile HTML file %s: %v", file, err)
		}

		firstHeader, err := GetFirstHeader("h1", htmlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to get first header for file %s: %v", file, err)
		}

		firstParagraph, err := GetFirstParagraph(htmlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to get first paragraph for file %s: %v", file, err)
		}

		markdownContent, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read markdown content for file %s: %v", file, err)
		}
		metadata, err := GetMetadata(string(markdownContent))
		if err != nil {
			logger.Warn("failed to get metadata for file %s: %v", file, err)
			metadata = nil
		}

		var lastModifiedDate time.Time
		fileModifiedDate, err := files.GetFileModifiedDate(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get modified date for file %s: %v", file, err)
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
		sortedSiteMap, err := SortSiteMap(siteMap, siteConfig.Site.SortDirection)
		if err != nil {
			return nil, fmt.Errorf("failed to sort site map: %v", err)
		}

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
	err = os.WriteFile(filePath, jsonContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to save site map to %s: %v", filePath, err)
	}
	return nil
}

func LoadSiteMap(filePath string) (*[]SiteMapEntry, error) {
	jsonContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var siteMap []SiteMapEntry
	err = json.Unmarshal(jsonContent, &siteMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal site map from %s: %v", filePath, err)
	}

	return &siteMap, nil
}

func FilterSiteMap(siteMap *[]SiteMapEntry, regex string) (*[]SiteMapEntry, error) {
	regexObj, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %v", err)
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
