package repocard

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestRepoCardExtension(t *testing.T) {
	tests := []struct {
		name             string
		markdown         string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "Valid repo card",
			markdown: `:::repo
jaysongiroux/mdserve
:::`,
			shouldContain: []string{
				`class="repo-card"`,
				`https://gh-card.dev/repos/jaysongiroux/mdserve.svg`,
				`https://github.com/jaysongiroux/mdserve`,
				`alt="jaysongiroux/mdserve"`,
			},
			shouldNotContain: []string{
				`:::repo`,
				`repo-card-error`,
			},
		},
		{
			name: "Repo card with spaces",
			markdown: `:::repo
  facebook/react  
:::`,
			shouldContain: []string{
				`https://gh-card.dev/repos/facebook/react.svg`,
				`https://github.com/facebook/react`,
			},
			shouldNotContain: []string{
				`repo-card-error`,
			},
		},
		{
			name: "Invalid repo format (no slash)",
			markdown: `:::repo
invalidrepo
:::`,
			shouldContain: []string{
				`repo-card-error`,
				`Invalid repository format`,
			},
			shouldNotContain: []string{
				`gh-card.dev`,
			},
		},
		{
			name: "Empty repo card",
			markdown: `:::repo
:::`,
			shouldContain: []string{
				`repo-card-error`,
			},
			shouldNotContain: []string{
				`gh-card.dev`,
			},
		},
		{
			name: "Repo card with other content",
			markdown: `Some text before

:::repo
vercel/next.js
:::

Some text after`,
			shouldContain: []string{
				`Some text before`,
				`https://gh-card.dev/repos/vercel/next.js.svg`,
				`Some text after`,
			},
			shouldNotContain: []string{
				`:::repo`,
			},
		},
		{
			name: "Multiple repo cards",
			markdown: `:::repo
facebook/react
:::

:::repo
vuejs/vue
:::`,
			shouldContain: []string{
				`https://gh-card.dev/repos/facebook/react.svg`,
				`https://gh-card.dev/repos/vuejs/vue.svg`,
			},
			shouldNotContain: []string{
				`repo-card-error`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(
					RepoCard,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.markdown), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()

			t.Logf("Markdown input:\n%s\n", tt.markdown)
			t.Logf("HTML output:\n%s\n", output)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
				}
			}

			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(output, unexpected) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nFull output:\n%s", unexpected, output)
				}
			}
		})
	}
}

func TestRepoCardValidation(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		isValid  bool
	}{
		{
			name:     "Valid standard repo",
			markdown: ":::repo\nuser/repo\n:::",
			isValid:  true,
		},
		{
			name:     "Valid with hyphens",
			markdown: ":::repo\nmy-user/my-repo\n:::",
			isValid:  true,
		},
		{
			name:     "Valid with underscores",
			markdown: ":::repo\nmy_user/my_repo\n:::",
			isValid:  true,
		},
		{
			name:     "Valid with numbers",
			markdown: ":::repo\nuser123/repo456\n:::",
			isValid:  true,
		},
		{
			name:     "Invalid with special chars",
			markdown: ":::repo\nuser@name/repo\n:::",
			isValid:  false,
		},
		{
			name:     "Invalid missing slash",
			markdown: ":::repo\nusername\n:::",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(RepoCard),
				goldmark.WithRendererOptions(html.WithUnsafe()),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.markdown), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()
			hasError := strings.Contains(output, "repo-card-error")

			if tt.isValid && hasError {
				t.Errorf("Expected valid repo card, but got error.\nOutput: %s", output)
			}
			if !tt.isValid && !hasError {
				t.Errorf("Expected error for invalid repo card, but got valid output.\nOutput: %s", output)
			}
		})
	}
}
