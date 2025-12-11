// Supports "captions" in markdown where it text align and italicized the text in a paragraph
// notation:
// ^^ this is a caption ^^
// ^^this is a caption^^
package caption

import (
	"bytes"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindCaption = gast.NewNodeKind("Caption")

type CaptionNode struct {
	gast.BaseInline
}

func (n *CaptionNode) Kind() gast.NodeKind {
	return KindCaption
}

func (n *CaptionNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

func NewCaption() *CaptionNode {
	return &CaptionNode{}
}

type captionParser struct {
}

var defaultCaptionParser = &captionParser{}

func NewCaptionParser() parser.InlineParser {
	return defaultCaptionParser
}

func (s *captionParser) Trigger() []byte {
	return []byte{'^'}
}

func (s *captionParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, segment := block.PeekLine()

	// minimum caption is ^^^^ (4 characters)
	if len(line) < 4 || line[0] != '^' || line[1] != '^' {
		return nil
	}

	closingPos := -1
	for i := 2; i < len(line)-1; i++ {
		if line[i] == '^' && line[i+1] == '^' {
			closingPos = i
			break
		}
	}
	if closingPos == -1 {
		return nil
	}

	content := line[2:closingPos]
	content = bytes.TrimSpace(content)

	if len(content) == 0 {
		return nil
	}

	node := NewCaption()

	contentStart := segment.Start + 2
	for contentStart < segment.Start+closingPos && line[contentStart-segment.Start] == ' ' {
		contentStart++
	}

	contentEnd := segment.Start + closingPos
	for contentEnd > contentStart && line[contentEnd-segment.Start-1] == ' ' {
		contentEnd--
	}

	textSegment := text.NewSegment(contentStart, contentEnd)
	textNode := gast.NewTextSegment(textSegment)
	node.AppendChild(node, textNode)
	block.Advance(closingPos + 2)

	return node
}

func (s *captionParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

type CaptionHTMLRenderer struct {
	html.Config
}

func NewCaptionHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &CaptionHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *CaptionHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCaption, r.renderCaption)
}

var CaptionAttributeFilter = html.GlobalAttributeFilter

func (r *CaptionHTMLRenderer) renderCaption(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<span style=\"display: block; width: 100%; text-align: center; font-style: italic; opacity: 0.8;\"")
			html.RenderAttributes(w, n, CaptionAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<span style=\"display: block; width: 100%; text-align: center; font-style: italic; opacity: 0.8;\">")
		}
	} else {
		_, _ = w.WriteString("</span>")
	}
	return gast.WalkContinue, nil
}

type caption struct {
}

var Caption = &caption{}

func (e *caption) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewCaptionParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewCaptionHTMLRenderer(), 500),
	))
}
