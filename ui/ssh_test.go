package ui

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	goorg "github.com/niklasfasching/go-org/org"
)

// TestStylesWithForcedProfile tests that styles work with forced TrueColor
func TestStylesWithForcedProfile(t *testing.T) {
	// Create renderer with forced TrueColor (like we do in SSH handler)
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)

	styles := NewStyles(r)

	tests := []struct {
		name      string
		style     lipgloss.Style
		input     string
		wantCode  string
		checkType string
	}{
		{"Bold", styles.Bold, "test", "\x1b[1", "bold code"},
		{"Italic", styles.Italic, "test", "\x1b[3", "italic code"},
		{"Underline", styles.Underline, "test", "\x1b[4", "underline code"},
		{"Heading1", styles.Heading1, "test", "\x1b[1", "bold code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style.Render(tt.input)
			t.Logf("%s output: %q", tt.name, result)

			if !strings.Contains(result, tt.wantCode) {
				t.Errorf("%s: expected %s (%s), not found in output", tt.name, tt.checkType, tt.wantCode)
			}
		})
	}
}

// TestRendererColorProfile verifies the renderer color profile is set correctly
func TestRendererColorProfile(t *testing.T) {
	r := lipgloss.NewRenderer(os.Stdout)

	// Before setting profile
	bold1 := r.NewStyle().Bold(true).Render("test")
	t.Logf("Before SetColorProfile - Bold: %q", bold1)

	// After setting TrueColor
	r.SetColorProfile(termenv.TrueColor)
	bold2 := r.NewStyle().Bold(true).Render("test")
	t.Logf("After SetColorProfile(TrueColor) - Bold: %q", bold2)

	if !strings.Contains(bold2, "\x1b[1") {
		t.Error("Bold style should contain \\x1b[1 after forcing TrueColor")
	}
}

// TestNewStylesPreservesRenderer tests that NewStyles uses the renderer correctly
func TestNewStylesPreservesRenderer(t *testing.T) {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)

	// Create styles AFTER setting color profile
	styles := NewStyles(r)

	boldResult := styles.Bold.Render("test")
	t.Logf("styles.Bold.Render: %q", boldResult)

	if !strings.Contains(boldResult, "\x1b[1") {
		t.Error("styles.Bold should produce bold ANSI code")
	}

	italicResult := styles.Italic.Render("test")
	t.Logf("styles.Italic.Render: %q", italicResult)

	if !strings.Contains(italicResult, "\x1b[3") {
		t.Error("styles.Italic should produce italic ANSI code")
	}
}

// TestMakeRendererSimulation simulates what bubbletea.MakeRenderer does
func TestMakeRendererSimulation(t *testing.T) {
	// This simulates the SSH session scenario
	var buf bytes.Buffer

	// Create renderer writing to buffer (like SSH session would)
	r := lipgloss.NewRenderer(&buf)

	t.Logf("Initial color profile: %v", r.ColorProfile())

	// Force TrueColor like we do in main.go
	r.SetColorProfile(termenv.TrueColor)
	t.Logf("After forcing TrueColor: %v", r.ColorProfile())

	// Create styles
	styles := NewStyles(r)

	// Render something
	output := styles.Bold.Render("bold") + " and " + styles.Italic.Render("italic")
	t.Logf("Rendered output: %q", output)

	// Verify ANSI codes
	if !strings.Contains(output, "\x1b[1") {
		t.Error("Missing bold ANSI code")
	}
	if !strings.Contains(output, "\x1b[3") {
		t.Error("Missing italic ANSI code")
	}
}

// TestEndToEndWithContext tests the full pipeline with context
func TestEndToEndWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate SSH output buffer
	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)

	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	// Parse and render org content
	input := `This has *bold* and /italic/ text.`

	// Use go-org to parse
	config := goorg.New()
	doc := config.Parse(strings.NewReader(input), "test.org")

	output := renderer.RenderNodes(doc.Nodes)

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	default:
		t.Logf("End-to-end output: %q", output)

		if !strings.Contains(output, "\x1b[1") {
			t.Error("Missing bold in end-to-end test")
		}
		if !strings.Contains(output, "\x1b[3") {
			t.Error("Missing italic in end-to-end test")
		}
	}
}

// TestSimulateSSHSession simulates what happens in an SSH session
// by creating a renderer that writes to a buffer (like SSH would)
func TestSimulateSSHSession(t *testing.T) {
	// Simulate the SSH session's output buffer
	var sessionOutput bytes.Buffer

	// This mimics what bubbletea.MakeRenderer(sess) does internally
	// It creates a renderer that writes to the session's stdout
	renderer := lipgloss.NewRenderer(&sessionOutput)

	// This mimics what we do in main.go: force TrueColor
	renderer.SetColorProfile(termenv.TrueColor)

	// Create styles using this renderer (like NewModel does)
	styles := NewStyles(renderer)

	// Create our renderer (like renderDocument does)
	orgRenderer := NewRenderer(styles, 80)

	// Parse some org content
	orgContent := `#+TITLE: Test

* Heading with *bold* word

This paragraph has *bold*, /italic/, ~code~, and =verbatim= text.

** Subheading

More /italic/ content here.
`

	config := goorg.New()
	doc := config.Parse(strings.NewReader(orgContent), "test.org")

	// Render the document
	output := orgRenderer.RenderNodes(doc.Nodes)

	t.Log("=== Simulated SSH Session Output ===")
	t.Log(output)
	t.Log("=== Raw Output ===")
	t.Logf("%q", output)

	// Check for expected ANSI codes
	checks := []struct {
		name     string
		code     string
		required bool
	}{
		{"bold (ESC[1)", "\x1b[1", true},
		{"italic (ESC[3)", "\x1b[3", true},
		{"color (ESC[38)", "\x1b[38", true},
		{"background (ESC[48)", "\x1b[48", false}, // code blocks have bg
	}

	for _, check := range checks {
		found := strings.Contains(output, check.code)
		if check.required && !found {
			t.Errorf("Missing required %s ANSI code in output", check.name)
		}
		if found {
			t.Logf("✓ Found %s", check.name)
		} else {
			t.Logf("✗ Not found: %s", check.name)
		}
	}

	// Count occurrences of bold
	boldCount := strings.Count(output, "\x1b[1")
	t.Logf("Bold code appears %d times", boldCount)

	if boldCount < 2 {
		t.Error("Expected at least 2 bold codes (heading + inline bold)")
	}
}

// TestInlineFormattingInParagraph specifically tests inline formatting
func TestInlineFormattingInParagraph(t *testing.T) {
	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)

	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	// Test just a paragraph with inline formatting
	input := `Text with *bold* and /italic/ words.`

	config := goorg.New()
	doc := config.Parse(strings.NewReader(input), "test.org")

	t.Log("=== Parsed Document Structure ===")
	for i, node := range doc.Nodes {
		t.Logf("Node %d: %T", i, node)
		if p, ok := node.(goorg.Paragraph); ok {
			for j, child := range p.Children {
				t.Logf("  Child %d: %T = %v", j, child, child)
			}
		}
	}

	output := renderer.RenderNodes(doc.Nodes)

	t.Log("=== Rendered Output ===")
	t.Log(output)
	t.Log("=== Raw ===")
	t.Logf("%q", output)

	// Detailed check
	if !strings.Contains(output, "\x1b[1") {
		t.Error("Bold formatting not found")
		t.Log("Expected \\x1b[1 (bold ANSI code) in output")
	}

	if !strings.Contains(output, "\x1b[3") {
		t.Error("Italic formatting not found")
		t.Log("Expected \\x1b[3 (italic ANSI code) in output")
	}
}

// TestDebugEmphasisRendering traces through the emphasis rendering
func TestDebugEmphasisRendering(t *testing.T) {
	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)

	styles := NewStyles(r)

	// Test the style directly first
	directBold := styles.Bold.Render("direct")
	t.Logf("Direct Bold.Render: %q", directBold)

	// Now test through the renderer
	renderer := NewRenderer(styles, 80)

	// Create an emphasis node manually
	emphNode := goorg.Emphasis{
		Kind:    "*",
		Content: []goorg.Node{goorg.Text{Content: "manual"}},
	}

	result := renderer.renderEmphasis(emphNode)
	t.Logf("renderEmphasis result: %q", result)

	// Test renderInlineNode
	inlineResult := renderer.renderInlineNode(emphNode)
	t.Logf("renderInlineNode result: %q", inlineResult)

	if directBold != result {
		t.Log("Warning: Direct style and renderEmphasis produce different results")
		t.Logf("  Direct: %q", directBold)
		t.Logf("  Via renderer: %q", result)
	}
}
