// Supports "github quote blocks" in markdown where it displays a quote block with a github like style
// Notation:
// > this is a plain quote block

// > [!NOTE]
// > Useful information that users should know, even when skimming content.

// > [!TIP]
// > Helpful advice for doing things better or more easily.

// > [!IMPORTANT]
// > Key information users need to know to achieve their goal.

// > [!WARNING]
// > Urgent info that needs immediate user attention to avoid problems.

// > [!CAUTION]
// > Advises about risks or negative outcomes of certain actions.
package githubquoteblock

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/jaysongiroux/mdserve/internal/logger"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Alert types
const (
	AlertTypeNote      = "NOTE"
	AlertTypeTip       = "TIP"
	AlertTypeImportant = "IMPORTANT"
	AlertTypeWarning   = "WARNING"
	AlertTypeCaution   = "CAUTION"
)

// Alert node kind
var KindAlert = gast.NewNodeKind("GitHubAlert")

// AlertNode represents a GitHub-style alert blockquote
type AlertNode struct {
	gast.BaseBlock
	AlertType string
}

// Kind returns the node kind
func (n *AlertNode) Kind() gast.NodeKind {
	return KindAlert
}

// Dump dumps the node to stdout
func (n *AlertNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{
		"AlertType": n.AlertType,
	}, nil)
}

// NewAlertNode creates a new alert node
func NewAlertNode(alertType string) *AlertNode {
	return &AlertNode{
		AlertType: alertType,
	}
}

// alertTransformer transforms blockquotes into alert nodes
type alertTransformer struct{}

var defaultAlertTransformer = &alertTransformer{}

// Transform transforms the AST
func (t *alertTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	// Pattern to match [!TYPE]
	alertPattern := regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]`)

	// Collect all blockquotes that need to be transformed
	type blockquoteTransform struct {
		blockquote *gast.Blockquote
		alertType  string
	}
	var transforms []blockquoteTransform

	// Walk through all nodes to find alert blockquotes
	_ = gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		// Only process blockquote nodes
		blockquote, ok := n.(*gast.Blockquote)
		if !ok {
			return gast.WalkContinue, nil
		}

		// Check if first child is a paragraph
		firstChild := blockquote.FirstChild()
		if firstChild == nil || firstChild.Kind() != gast.KindParagraph {
			return gast.WalkContinue, nil
		}

		paragraph := firstChild.(*gast.Paragraph)

		// Concatenate all text nodes in the paragraph to get the full text
		// Goldmark may split text like "[!NOTE]" into multiple text nodes
		var fullText bytes.Buffer
		for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
			if child.Kind() == gast.KindText {
				textNode := child.(*gast.Text)
				fullText.Write(textNode.Segment.Value(reader.Source()))
			}
		}

		text := fullText.Bytes()
		logger.Debug("Checking blockquote text: %s", string(text))

		// Check if it matches alert pattern
		matches := alertPattern.FindSubmatch(text)
		if len(matches) < 2 {
			logger.Debug("No alert pattern found in blockquote")
			return gast.WalkContinue, nil
		}

		logger.Debug("Matches: %s", string(matches[1]))
		alertType := string(matches[1])
		logger.Debug("Found GitHub alert blockquote: [!%s]", alertType)

		// Add to transforms list
		transforms = append(transforms, blockquoteTransform{
			blockquote: blockquote,
			alertType:  alertType,
		})

		return gast.WalkContinue, nil
	})

	// Now perform all transformations
	for _, transform := range transforms {
		blockquote := transform.blockquote
		alertType := transform.alertType

		// Create new alert node
		alert := NewAlertNode(alertType)

		// Move all children from blockquote to alert
		for child := blockquote.FirstChild(); child != nil; {
			nextChild := child.NextSibling()
			blockquote.RemoveChild(blockquote, child)
			alert.AppendChild(alert, child)
			child = nextChild
		}

		// Replace blockquote with alert node
		parent := blockquote.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, blockquote, alert)
			logger.Debug("Successfully transformed blockquote to %s alert", alertType)
		}
	}
}

// AlertHTMLRenderer renders alert nodes to HTML
type AlertHTMLRenderer struct {
	html.Config
	alertPattern *regexp.Regexp
}

// NewAlertHTMLRenderer creates a new alert renderer
func NewAlertHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &AlertHTMLRenderer{
		Config:       html.NewConfig(),
		alertPattern: regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*`),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers rendering functions
func (r *AlertHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAlert, r.renderAlert)
	reg.Register(gast.KindBlockquote, r.renderBlockquote)
}

// Alert configuration
type alertConfig struct {
	emoji       string
	title       string
	colorLight  string
	colorDark   string
	borderColor string
}

var alertConfigs = map[string]alertConfig{
	AlertTypeNote: {
		emoji:       "ℹ️",
		title:       "Note",
		colorLight:  "#ddf4ff",
		colorDark:   "#0969da",
		borderColor: "#54aeff",
	},
	AlertTypeTip: {
		emoji:       "💡",
		title:       "Tip",
		colorLight:  "#d8f5e3",
		colorDark:   "#1a7f37",
		borderColor: "#4ac776",
	},
	AlertTypeImportant: {
		emoji:       "⚠️",
		title:       "Important",
		colorLight:  "#f0e6ff",
		colorDark:   "#8250df",
		borderColor: "#a371f7",
	},
	AlertTypeWarning: {
		emoji:       "⚡",
		title:       "Warning",
		colorLight:  "#fff8e6",
		colorDark:   "#9a6700",
		borderColor: "#d4a72c",
	},
	AlertTypeCaution: {
		emoji:       "🔥",
		title:       "Caution",
		colorLight:  "#ffebe9",
		colorDark:   "#d1242f",
		borderColor: "#ff6369",
	},
}

// renderAlert renders an alert node
func (r *AlertHTMLRenderer) renderAlert(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
		alert := node.(*AlertNode)
		config, ok := alertConfigs[alert.AlertType]
		if !ok {
			// Fallback to note style
			config = alertConfigs[AlertTypeNote]
		}

		// Write opening div with styling
		_, _ = w.WriteString(`<div class="github-alert github-alert-`)
		_, _ = w.WriteString(strings.ToLower(alert.AlertType))
		_, _ = w.WriteString(`" style="`)
		_, _ = w.WriteString(`padding: 12px 16px; `)
		_, _ = w.WriteString(`margin-bottom: 16px; `)
		_, _ = w.WriteString(`border-left: 4px solid `)
		_, _ = w.WriteString(config.borderColor)
		_, _ = w.WriteString(`; `)
		_, _ = w.WriteString(`background-color: `)
		_, _ = w.WriteString(config.colorLight)
		_, _ = w.WriteString(`; `)
		_, _ = w.WriteString(`border-radius: 4px;`)
		_, _ = w.WriteString(`">`)

		// Write title with emoji
		_, _ = w.WriteString(`<div style="display: flex; align-items: center; margin-bottom: 8px; font-weight: 600; color: `)
		_, _ = w.WriteString(config.colorDark)
		_, _ = w.WriteString(`;">`)
		_, _ = w.WriteString(`<span style="margin-right: 8px;">`)
		_, _ = w.WriteString(config.emoji)
		_, _ = w.WriteString(`</span>`)
		_, _ = w.WriteString(config.title)
		_, _ = w.WriteString(`</div>`)

		// Write content wrapper
		_, _ = w.WriteString(`<div class="github-alert-content" style="color: `)
		_, _ = w.WriteString(config.colorDark)
		_, _ = w.WriteString(`;">`)

		// Render children, but strip [!TYPE] from first paragraph
		firstParagraph := true
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if child.Kind() == gast.KindParagraph {
				_, _ = w.WriteString(`<p style="margin-bottom: 8px;">`)

				if firstParagraph {
					// Concatenate all text nodes in the first paragraph
					// Goldmark may split text like "[!NOTE]" into multiple text nodes
					var fullText bytes.Buffer
					for grandchild := child.FirstChild(); grandchild != nil; grandchild = grandchild.NextSibling() {
						if grandchild.Kind() == gast.KindText {
							textNode := grandchild.(*gast.Text)
							fullText.Write(textNode.Segment.Value(source))
						}
					}

					// Strip [!TYPE] from the concatenated text
					cleanedText := r.alertPattern.ReplaceAll(fullText.Bytes(), []byte{})
					cleanedText = bytes.TrimSpace(cleanedText)
					_, _ = w.Write(cleanedText)
					firstParagraph = false
				} else {
					// Render other paragraphs normally
					for grandchild := child.FirstChild(); grandchild != nil; grandchild = grandchild.NextSibling() {
						if grandchild.Kind() == gast.KindText {
							segment := grandchild.(*gast.Text).Segment
							_, _ = w.Write(segment.Value(source))
						}
					}
				}

				_, _ = w.WriteString(`</p>`)
			}
		}
	} else {
		// Close content wrapper and main div
		_, _ = w.WriteString(`</div></div>`)
	}

	return gast.WalkSkipChildren, nil // Skip children since we already rendered them
}

// renderBlockquote renders plain blockquote nodes with custom styling
func (r *AlertHTMLRenderer) renderBlockquote(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
		// Write opening blockquote with gray styling
		_, _ = w.WriteString(`<blockquote style="`)
		_, _ = w.WriteString(`padding: 12px 16px; `)
		_, _ = w.WriteString(`margin: 16px 0; `)
		_, _ = w.WriteString(`border-left: 4px solid #d0d7de; `)
		_, _ = w.WriteString(`background-color: #f6f8fa; `)
		_, _ = w.WriteString(`border-radius: 4px; `)
		_, _ = w.WriteString(`color: #57606a;`)
		_, _ = w.WriteString(`">`)
		_, _ = w.WriteString("\n")

		// Render children with custom paragraph styling
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if child.Kind() == gast.KindParagraph {
				_, _ = w.WriteString(`<p style="margin-bottom: 8px;">`)
				// Render all text nodes in the paragraph
				for grandchild := child.FirstChild(); grandchild != nil; grandchild = grandchild.NextSibling() {
					if grandchild.Kind() == gast.KindText {
						textNode := grandchild.(*gast.Text)
						_, _ = w.Write(textNode.Segment.Value(source))
					}
				}
				_, _ = w.WriteString(`</p>`)
				_, _ = w.WriteString("\n")
			}
		}
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}

	return gast.WalkSkipChildren, nil
}

// gitHubAlertExtension is the extension struct
type gitHubAlertExtension struct{}

// GitHubQuoteBlock is the extension instance
var GitHubQuoteBlock = &gitHubAlertExtension{}

// Extend extends the Goldmark processor
func (e *gitHubAlertExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(defaultAlertTransformer, 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewAlertHTMLRenderer(), 500),
		),
	)
}
