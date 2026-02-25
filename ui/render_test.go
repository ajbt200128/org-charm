package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	goorg "github.com/niklasfasching/go-org/org"
)

// createTestRenderer creates a renderer with forced TrueColor like SSH would have
func createTestRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)
	return r
}

func TestInlineEmphasis(t *testing.T) {
	r := createTestRenderer()
	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	tests := []struct {
		name     string
		input    string
		wantBold bool
		wantItal bool
		wantCode bool
	}{
		{
			name:     "bold text",
			input:    "This has *bold* text.",
			wantBold: true,
		},
		{
			name:     "italic text",
			input:    "This has /italic/ text.",
			wantItal: true,
		},
		{
			name:     "code text",
			input:    "This has ~code~ text.",
			wantCode: true,
		},
		{
			name:     "mixed formatting",
			input:    "This has *bold* and /italic/ and ~code~ text.",
			wantBold: true,
			wantItal: true,
			wantCode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := goorg.New()
			doc := config.Parse(strings.NewReader(tt.input), "test.org")

			output := renderer.RenderNodes(doc.Nodes)

			t.Logf("Input: %s", tt.input)
			t.Logf("Output: %s", output)
			t.Logf("Raw output: %q", output)

			// Check for ANSI bold code (ESC[1m)
			if tt.wantBold && !strings.Contains(output, "\x1b[1") {
				t.Errorf("Expected bold ANSI code (\\x1b[1), not found in output")
			}

			// Check for ANSI italic code (ESC[3m)
			if tt.wantItal && !strings.Contains(output, "\x1b[3") {
				t.Errorf("Expected italic ANSI code (\\x1b[3), not found in output")
			}

			// Check for background color (code blocks have background)
			if tt.wantCode && !strings.Contains(output, "\x1b[") {
				t.Errorf("Expected ANSI codes for code, not found in output")
			}
		})
	}
}

func TestEmphasisNodeParsing(t *testing.T) {
	input := "Text with *bold* and /italic/ words."
	config := goorg.New()
	doc := config.Parse(strings.NewReader(input), "test.org")

	t.Log("=== Parsed nodes ===")
	for _, node := range doc.Nodes {
		logNode(t, node, 0)
	}
}

func logNode(t *testing.T, node goorg.Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case goorg.Paragraph:
		t.Logf("%sParagraph", indent)
		for _, child := range n.Children {
			logNode(t, child, depth+1)
		}
	case goorg.Text:
		t.Logf("%sText: %q", indent, n.Content)
	case goorg.Emphasis:
		t.Logf("%sEmphasis(Kind=%q)", indent, n.Kind)
		for _, child := range n.Content {
			logNode(t, child, depth+1)
		}
	default:
		t.Logf("%s%T: %v", indent, node, node)
	}
}

func TestRenderEmphasisDirectly(t *testing.T) {
	r := createTestRenderer()
	styles := NewStyles(r)

	// Test styles directly
	boldResult := styles.Bold.Render("bold")
	italicResult := styles.Italic.Render("italic")
	codeResult := styles.InlineCode.Render("code")

	t.Logf("Bold style output: %q", boldResult)
	t.Logf("Italic style output: %q", italicResult)
	t.Logf("Code style output: %q", codeResult)

	if !strings.Contains(boldResult, "\x1b[1") {
		t.Errorf("Bold style doesn't contain bold ANSI code")
	}
	if !strings.Contains(italicResult, "\x1b[3") {
		t.Errorf("Italic style doesn't contain italic ANSI code")
	}
}

func TestFullRenderPipeline(t *testing.T) {
	r := createTestRenderer()
	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	input := `#+TITLE: Test

* Heading

This paragraph has *bold*, /italic/, and ~code~ formatting.

Another paragraph with _underline_ and +strikethrough+ text.
`

	config := goorg.New()
	doc := config.Parse(strings.NewReader(input), "test.org")

	output := renderer.RenderNodes(doc.Nodes)

	t.Log("=== Full render output ===")
	t.Log(output)
	t.Log("=== Raw output ===")
	t.Logf("%q", output)

	// Verify ANSI codes are present
	checks := []struct {
		name string
		code string
	}{
		{"bold", "\x1b[1"},
		{"italic", "\x1b[3"},
	}

	for _, check := range checks {
		if !strings.Contains(output, check.code) {
			t.Errorf("Missing %s ANSI code in output", check.name)
		}
	}
}

func TestRenderInlineNodeSwitch(t *testing.T) {
	r := createTestRenderer()
	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	// Create emphasis node manually
	emphasisNode := goorg.Emphasis{
		Kind: "*",
		Content: []goorg.Node{
			goorg.Text{Content: "bold text"},
		},
	}

	result := renderer.renderInlineNode(emphasisNode)
	t.Logf("Direct renderInlineNode result: %q", result)

	if !strings.Contains(result, "\x1b[1") {
		t.Errorf("renderInlineNode didn't produce bold ANSI code")
	}
}
