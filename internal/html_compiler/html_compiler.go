package htmlcompiler

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/html_compiler/extention/caption"
	githubquoteblock "github.com/jaysongiroux/mdserve/internal/html_compiler/extention/github_quoteblock"
	"github.com/jaysongiroux/mdserve/internal/logger"
	"github.com/jaysongiroux/mdserve/internal/routines"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"golang.org/x/sync/errgroup"
)

type SiteMapEntry struct {
	Path             string    `json:"path"`
	Metadata         *Metadata `json:"metadata"`
	FirstHeader      string    `json:"first_header"`
	FirstParagraph   string    `json:"first_paragraph"`
	LastModifiedDate time.Time `json:"last_modified_date"`
	CreationDate     time.Time `json:"creation_date"`
}

func CompileHTMLFiles(
	mdFiles []string,
	siteConfig *config.SiteConfig,
	serverConfig *config.ServerConfig,
) error {
	if len(mdFiles) == 0 {
		logger.Warn("No MD files to compile, skipping HTML compilation")
		return nil
	}

	maxWorkers := routines.CalculateMaxWorkers(len(mdFiles))
	logger.Info("Compiling %d MD files with %d concurrent workers", len(mdFiles), maxWorkers)

	g := new(errgroup.Group)
	g.SetLimit(maxWorkers)

	for i, mdFile := range mdFiles {
		g.Go(func() error {
			logger.Info("[Worker %d] Compiling MD file: %s", i, mdFile)
			htmlFile, err := CompileHTMLFile(mdFile, siteConfig)
			logger.Debug("[Worker %d] Compiling HTML file: %s", i, mdFile)
			if err != nil {
				logger.Fatal("[Worker %d] Failed to compile HTML file: %v", i, err)
				return fmt.Errorf("failed to compile HTML file: %w", err)
			}

			// Calculate relative path from content directory to preserve folder structure
			relPath, err := filepath.Rel(serverConfig.ContentPath, mdFile)
			if err != nil {
				logger.Fatal("[Worker %d] Failed to get relative path for %s: %v", i, mdFile, err)
				return fmt.Errorf("failed to get relative path for %s: %w", mdFile, err)
			}

			// save the HTML file to the compiled HTML path
			// write files to the content path /.html/ preserving directory structure
			savePath := filepath.Join(serverConfig.GeneratedPath)
			logger.Debug("[Worker %d] Writing HTML file: %s to %s", i, relPath, savePath)

			err = WriteHTMLFile(savePath, relPath, htmlFile)
			if err != nil {
				logger.Fatal("[Worker %d] Failed to write HTML file: %v", i, err)
				return fmt.Errorf("failed to write HTML file: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to compile HTML files: %w", err)
	}

	logger.Info("Successfully compiled all MD files")
	return nil
}

// CompileHTMLFile converts a markdown file to an HTML string
func CompileHTMLFile(filePath string, siteConfig *config.SiteConfig) (string, error) {
	// 1. Read the file from disk
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			githubquoteblock.GitHubQuoteBlock,
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			extension.CJK,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
			caption.Caption,
			&mermaid.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle(siteConfig.Site.Theme.Code.Theme),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(siteConfig.Site.Theme.Code.LineNumbers),
					chromahtml.BaseLineNumber(0),
					chromahtml.InlineCode(true),
					chromahtml.WrapLongLines(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// 2. Convert the byte slice to HTML
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	// 3. Replace asset paths with generated assets paths
	htmlContent := buf.String()
	htmlContent = replaceAssetPaths(htmlContent)

	return htmlContent, nil
}

// replaceAssetPaths replaces asset paths in HTML with generated assets paths
func replaceAssetPaths(htmlContent string) string {
	re := regexp.MustCompile(`(src|href)="([^"]*)"`)

	return re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		parts := regexp.MustCompile(`(src|href)="([^"]*)"`).FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		attrName := parts[1]
		attrValue := parts[2]

		if !strings.Contains(attrValue, "/"+constants.GeneratedAssetsPath+"/") {
			return match
		}

		ext := strings.ToLower(filepath.Ext(attrValue))

		newPath := attrValue
		for _, optimizableExt := range constants.OptimizableImageExtensions {
			if ext == optimizableExt {
				newPath = strings.TrimSuffix(attrValue, ext) + ".webp"
				break
			}
		}

		return attrName + `="` + newPath + `"`
	})
}

func GetHeaders(htmlContent string) ([]Header, error) {
	var headers []Header

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return headers, err
	}

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		level := goquery.NodeName(s)
		text := s.Text()
		headers = append(headers, Header{Level: levelToInt(level), Text: text})
	})

	return headers, nil
}

func levelToInt(level string) int {
	switch level {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		log.Fatalf("Unexpected header level: %s", level)
	}
	return 0
}

func GetFirstHeader(headerType string, HTMLContent string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTMLContent))
	if err != nil {
		return "", err
	}
	header := doc.Find(headerType).First()
	if header.Length() == 0 {
		return "", errors.New("no header found")
	}
	return header.Text(), nil
}

func GetFirstParagraph(HTMLContent string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(HTMLContent))
	if err != nil {
		return "", err
	}

	var firstNonEmptyParagraph string
	doc.Find("p").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			firstNonEmptyParagraph = text
			return false
		}
		return true
	})

	if firstNonEmptyParagraph == "" {
		logger.Debug("No non-empty paragraph found in HTML content")
		return "", nil
	}

	return firstNonEmptyParagraph, nil
}
