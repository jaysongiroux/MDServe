package htmlcompiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Metadata struct {
	Tags                 []string  `json:"tags"`
	CreationDate         time.Time `json:"creation_date"`
	Description          string    `json:"description"`
	Author               string    `json:"author"`
	LastModificationDate time.Time `json:"last_modification_date"`
}

func GetMetadata(markdownContent string) (*Metadata, error) {
	// Check if the file starts with a comment that is a JSON object
	if !strings.HasPrefix(markdownContent, "<!--") {
		return nil, nil
	}

	// Get the contents of the comment
	startIndex := len("<!--")
	endIndex := strings.Index(markdownContent, "-->")
	if endIndex == -1 {
		return nil, errors.New("invalid metadata format: missing closing -->")
	}
	commentContents := markdownContent[startIndex:endIndex]

	// Trim whitespace around the JSON object
	jsonString := strings.TrimSpace(commentContents)

	// Parse the extracted JSON string into a map
	var metadata Metadata
	err := json.Unmarshal([]byte(jsonString), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &Metadata{
		Tags:                 metadata.Tags,
		CreationDate:         metadata.CreationDate,
		Description:          metadata.Description,
		Author:               metadata.Author,
		LastModificationDate: metadata.LastModificationDate,
	}, nil
}

func GetModifiedDate(siteMapEntry SiteMapEntry) time.Time {
	if siteMapEntry.Metadata != nil {
		if siteMapEntry.Metadata.LastModificationDate != (time.Time{}) {
			return siteMapEntry.Metadata.LastModificationDate
		}
	}
	return siteMapEntry.LastModifiedDate
}

func GetCreationDate(siteMapEntry SiteMapEntry) time.Time {
	if siteMapEntry.Metadata != nil {
		if siteMapEntry.Metadata.CreationDate != (time.Time{}) {
			return siteMapEntry.Metadata.CreationDate
		}
	}
	return siteMapEntry.CreationDate
}
