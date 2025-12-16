// Package repocard provides a Goldmark extension for GitHub repository cards
// Supports displaying repository information using gh-card.dev
//
// Notation:
// :::repo
// jaysongiroux/mdserve
// :::
package repocard

import (
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

var KindRepoCard = gast.NewNodeKind("RepoCard")

type RepoCardNode struct {
	gast.BaseBlock
	Owner string
	Repo  string
}

func (n *RepoCardNode) Kind() gast.NodeKind {
	return KindRepoCard
}

func (n *RepoCardNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, map[string]string{
		"Owner": n.Owner,
		"Repo":  n.Repo,
	}, nil)
}

type repoCardParser struct{}

func NewRepoCardParser() parser.BlockParser {
	return &repoCardParser{}
}

func (p *repoCardParser) Trigger() []byte {
	return []byte{':'}
}

func (p *repoCardParser) Open(
	parent gast.Node,
	reader text.Reader,
	pc parser.Context,
) (gast.Node, parser.State) {
	line, segment := reader.PeekLine()
	lineStr := strings.TrimSpace(string(line))

	if lineStr != ":::repo" {
		return nil, parser.NoChildren
	}

	// Advance past :::repo line
	reader.Advance(segment.Len())

	var owner, repo string

	// Read lines until we find closing :::
	for {
		line, segment := reader.PeekLine()
		if line == nil {
			break
		}

		lineStr := strings.TrimSpace(string(line))
		reader.Advance(segment.Len())

		if lineStr == ":::" {
			break
		}

		if lineStr == "" {
			continue
		}

		// Parse owner/repo if not already set
		if owner == "" && repo == "" {
			parts := strings.Split(lineStr, "/")
			if len(parts) == 2 {
				owner = strings.TrimSpace(parts[0])
				repo = strings.TrimSpace(parts[1])
			}
		}
	}

	return &RepoCardNode{Owner: owner, Repo: repo}, parser.Close
}

func (p *repoCardParser) Continue(
	node gast.Node,
	reader text.Reader,
	pc parser.Context,
) parser.State {
	return parser.Close
}

func (p *repoCardParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {}

func (p *repoCardParser) CanInterruptParagraph() bool {
	return true
}

func (p *repoCardParser) CanAcceptIndentedLine() bool {
	return false
}

type RepoCardHTMLRenderer struct {
	html.Config
}

func NewRepoCardHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &RepoCardHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *RepoCardHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindRepoCard, r.renderRepoCard)
}

var repoNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func (r *RepoCardHTMLRenderer) renderRepoCard(
	w util.BufWriter,
	source []byte,
	node gast.Node,
	entering bool,
) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}

	repoCard := node.(*RepoCardNode)

	if repoCard.Owner == "" || repoCard.Repo == "" {
		logger.Fatal("Invalid repo card: missing owner or repo name")
	}

	if !repoNamePattern.MatchString(repoCard.Owner) || !repoNamePattern.MatchString(repoCard.Repo) {
		logger.Fatal("Invalid repo card format: %s/%s", repoCard.Owner, repoCard.Repo)
	}

	cardURL := "https://gh-card.dev/repos/" + repoCard.Owner + "/" + repoCard.Repo + ".svg"
	repoURL := "https://github.com/" + repoCard.Owner + "/" + repoCard.Repo

	_, _ = w.WriteString(`<div class="repo-card" style="margin: 20px 0; max-width: 500px;">`)
	_, _ = w.WriteString(`<a href="`)
	_, _ = w.WriteString(repoURL)
	_, _ = w.WriteString(
		`" target="_blank" rel="noopener noreferrer" style="text-decoration: none;">`,
	)
	_, _ = w.WriteString(`<img src="`)
	_, _ = w.WriteString(cardURL)
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.WriteString(repoCard.Owner)
	_, _ = w.WriteString(`/`)
	_, _ = w.WriteString(repoCard.Repo)
	_, _ = w.WriteString(
		`" style="max-width: 100%; height: auto; border-radius: 6px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">`,
	)
	_, _ = w.WriteString(`</a>`)
	_, _ = w.WriteString(`</div>`)

	return gast.WalkContinue, nil
}

type repoCardExtension struct{}

var RepoCard = &repoCardExtension{}

func (e *repoCardExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewRepoCardParser(), 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewRepoCardHTMLRenderer(), 500),
		),
	)
}
