package githubquoteblock

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestGitHubAlertExtension(t *testing.T) {
	tests := []struct {
		name             string
		markdown         string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "NOTE alert",
			markdown: `> [!NOTE]
> This is a note alert`,
			shouldContain: []string{
				`class="github-alert github-alert-note"`,
				`ℹ️`,
				`Note`,
				`This is a note alert`,
				`background-color: #ddf4ff`,
				`border-left: 4px solid #54aeff`,
			},
			shouldNotContain: []string{
				`[!NOTE]`,
				`<blockquote>`,
			},
		},
		{
			name: "TIP alert",
			markdown: `> [!TIP]
> Here's a helpful tip`,
			shouldContain: []string{
				`class="github-alert github-alert-tip"`,
				`💡`,
				`Tip`,
				`Here's a helpful tip`,
				`background-color: #d8f5e3`,
			},
			shouldNotContain: []string{
				`[!TIP]`,
			},
		},
		{
			name: "IMPORTANT alert",
			markdown: `> [!IMPORTANT]
> This is important information`,
			shouldContain: []string{
				`class="github-alert github-alert-important"`,
				`⚠️`,
				`Important`,
				`This is important information`,
				`background-color: #f0e6ff`,
			},
			shouldNotContain: []string{
				`[!IMPORTANT]`,
			},
		},
		{
			name: "WARNING alert",
			markdown: `> [!WARNING]
> Be careful about this`,
			shouldContain: []string{
				`class="github-alert github-alert-warning"`,
				`⚡`,
				`Warning`,
				`Be careful about this`,
				`background-color: #fff8e6`,
			},
			shouldNotContain: []string{
				`[!WARNING]`,
			},
		},
		{
			name: "CAUTION alert",
			markdown: `> [!CAUTION]
> It is not recommended to use your password, instead use a personal access token.`,
			shouldContain: []string{
				`class="github-alert github-alert-caution"`,
				`🔥`,
				`Caution`,
				`It is not recommended to use your password, instead use a personal access token.`,
				`background-color: #ffebe9`,
				`border-left: 4px solid #ff6369`,
			},
			shouldNotContain: []string{
				`[!CAUTION]`,
				`<blockquote>`,
			},
		},
		{
			name: "Plain blockquote (no alert)",
			markdown: `> This is just a regular blockquote
> Nothing special here`,
			shouldContain: []string{
				`<blockquote`,
				`This is just a regular blockquote`,
			},
			shouldNotContain: []string{
				`class="github-alert`,
				`ℹ️`,
				`💡`,
			},
		},
		{
			name:     "Alert with inline on same line",
			markdown: `> [!NOTE] This is all on one line`,
			shouldContain: []string{
				`class="github-alert github-alert-note"`,
				`This is all on one line`,
			},
			shouldNotContain: []string{
				`[!NOTE]`,
			},
		},
		{
			name: "Multi-paragraph alert",
			markdown: `> [!TIP]
> First paragraph here.
>
> Second paragraph here.`,
			shouldContain: []string{
				`class="github-alert github-alert-tip"`,
				`First paragraph here`,
				`Second paragraph here`,
			},
			shouldNotContain: []string{
				`[!TIP]`,
			},
		},
		{
			name: "Multiple different alerts in a row",
			markdown: `> this is a plain quote block

> [!NOTE]
> Useful information that users should know, even when skimming content.

> [!TIP]
> Helpful advice for doing things better or more easily.`,
			shouldContain: []string{
				// Plain blockquote
				`<blockquote`,
				`this is a plain quote block`,
				// NOTE alert
				`class="github-alert github-alert-note"`,
				`Note`,
				`Useful information that users should know`,
				// TIP alert
				`class="github-alert github-alert-tip"`,
				`Tip`,
				`Helpful advice for doing things better`,
			},
			shouldNotContain: []string{
				`[!NOTE]`,
				`[!TIP]`,
			},
		},
		{
			name: "Plain blockquote with multiple lines",
			markdown: `> This is a plain blockquote
> with multiple lines
> of text`,
			shouldContain: []string{
				`<blockquote`,
				`This is a plain blockquote`,
				`with multiple lines`,
			},
			shouldNotContain: []string{
				`class="github-alert`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Goldmark instance with the extension
			md := goldmark.New(
				goldmark.WithExtensions(
					GitHubQuoteBlock,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)

			// Convert markdown to HTML
			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.markdown), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()

			// Check for expected strings
			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					// Print output for debugging
					t.Logf("Markdown input:\n%s\n", tt.markdown)
					t.Logf("HTML output:\n%s\n", output)
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
				}
			}

			// Check for unexpected strings
			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(output, unexpected) {
					t.Logf("Markdown input:\n%s\n", tt.markdown)
					t.Logf("HTML output:\n%s\n", output)
					t.Errorf("Expected output NOT to contain %q, but it did.\nFull output:\n%s", unexpected, output)
				}
			}
		})
	}
}

func TestAlertTransformerDetection(t *testing.T) {
	tests := []struct {
		name            string
		markdown        string
		shouldTransform bool
	}{
		{
			name:            "Valid NOTE",
			markdown:        "> [!NOTE]\n> Content",
			shouldTransform: true,
		},
		{
			name:            "Invalid - lowercase",
			markdown:        "> [!note]\n> Content",
			shouldTransform: false,
		},
		{
			name:            "Invalid - spaces",
			markdown:        "> [! NOTE ]\n> Content",
			shouldTransform: false,
		},
		{
			name:            "Invalid - unknown type",
			markdown:        "> [!UNKNOWN]\n> Content",
			shouldTransform: false,
		},
		{
			name:            "Plain blockquote",
			markdown:        "> Just a quote",
			shouldTransform: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithExtensions(GitHubQuoteBlock),
				goldmark.WithRendererOptions(html.WithUnsafe()),
			)

			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.markdown), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			output := buf.String()
			hasAlertClass := strings.Contains(output, `class="github-alert`)
			hasBlockquote := strings.Contains(output, `<blockquote>`)

			if tt.shouldTransform {
				if !hasAlertClass {
					t.Errorf("Expected markdown to be transformed to alert, but it wasn't.\nOutput: %s", output)
				}
				if hasBlockquote {
					t.Errorf("Expected blockquote to be replaced with alert, but blockquote still exists.\nOutput: %s", output)
				}
			} else {
				if hasAlertClass {
					t.Errorf("Expected markdown NOT to be transformed to alert, but it was.\nOutput: %s", output)
				}
			}
		})
	}
}

func BenchmarkGitHubAlert(b *testing.B) {
	markdown := `> [!NOTE]
> This is a note alert with some content
> that spans multiple lines`

	md := goldmark.New(
		goldmark.WithExtensions(GitHubQuoteBlock),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = md.Convert([]byte(markdown), &buf)
	}
}
