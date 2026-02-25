package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	goorg "github.com/niklasfasching/go-org/org"
)

// TestGenerateSnapshot creates a snapshot file of the rendered output
// Run with: go test -v ./ui/... -run TestGenerateSnapshot
// Then inspect: cat -v /tmp/org-charm-snapshot.txt
func TestGenerateSnapshot(t *testing.T) {
	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)

	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)

	orgContent := `#+TITLE: Snapshot Test
#+AUTHOR: Test

* Heading Level 1

This paragraph has *bold text*, /italic text/, ~inline code~, =verbatim=, _underline_, and +strikethrough+.

** Heading Level 2

- List item with *bold*
- List item with /italic/
- [X] Done checkbox
- [ ] Todo checkbox

*** Heading Level 3

#+BEGIN_QUOTE
This is a quoted block with /italic/ inside.
#+END_QUOTE

#+BEGIN_SRC go
func main() {
    fmt.Println("Hello")
}
#+END_SRC

| Column 1 | Column 2 |
|----------+----------|
| *bold*   | /italic/ |
`

	config := goorg.New()
	doc := config.Parse(strings.NewReader(orgContent), "test.org")

	output := renderer.RenderNodes(doc.Nodes)

	// Write to snapshot file
	snapshotPath := "/tmp/org-charm-snapshot.txt"
	err := os.WriteFile(snapshotPath, []byte(output), 0644)
	if err != nil {
		t.Fatalf("Failed to write snapshot: %v", err)
	}

	t.Logf("Snapshot written to: %s", snapshotPath)
	t.Logf("View with: cat %s", snapshotPath)
	t.Logf("View raw with: cat -v %s", snapshotPath)
	t.Logf("View hex with: xxd %s | head -100", snapshotPath)

	// Also log to test output
	t.Log("\n=== SNAPSHOT OUTPUT ===")
	t.Log(output)

	t.Log("\n=== RAW BYTES (first 500) ===")
	if len(output) > 500 {
		t.Logf("%q...", output[:500])
	} else {
		t.Logf("%q", output)
	}

	// Verify expected codes
	expectedCodes := map[string]string{
		"bold":      "\x1b[1",
		"italic":    "\x1b[3",
		"underline": "\x1b[4",
		"color":     "\x1b[38;2",
	}

	t.Log("\n=== ANSI CODE CHECK ===")
	for name, code := range expectedCodes {
		count := strings.Count(output, code)
		if count > 0 {
			t.Logf("✓ %s (%s): found %d times", name, code, count)
		} else {
			t.Errorf("✗ %s (%s): NOT FOUND", name, code)
		}
	}
}

// TestCompareWithAndWithoutColorProfile shows the difference
func TestCompareWithAndWithoutColorProfile(t *testing.T) {
	orgContent := `Text with *bold* and /italic/.`

	config := goorg.New()
	doc := config.Parse(strings.NewReader(orgContent), "test.org")

	t.Log("=== WITHOUT SetColorProfile ===")
	{
		var buf bytes.Buffer
		r := lipgloss.NewRenderer(&buf)
		// NOT setting color profile - this is what might happen if detection fails
		styles := NewStyles(r)
		renderer := NewRenderer(styles, 80)
		output := renderer.RenderNodes(doc.Nodes)
		t.Logf("Output: %s", output)
		t.Logf("Raw: %q", output)
		t.Logf("Has bold code: %v", strings.Contains(output, "\x1b[1"))
	}

	t.Log("\n=== WITH SetColorProfile(TrueColor) ===")
	{
		var buf bytes.Buffer
		r := lipgloss.NewRenderer(&buf)
		r.SetColorProfile(termenv.TrueColor)
		styles := NewStyles(r)
		renderer := NewRenderer(styles, 80)
		output := renderer.RenderNodes(doc.Nodes)
		t.Logf("Output: %s", output)
		t.Logf("Raw: %q", output)
		t.Logf("Has bold code: %v", strings.Contains(output, "\x1b[1"))
	}
}

// TestModelView tests what Model.View() actually produces
func TestModelView(t *testing.T) {
	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)

	// Parse test files
	orgContent := `#+TITLE: Test File

* Section with *bold*

Paragraph with /italic/ text.
`
	config := goorg.New()
	doc := config.Parse(strings.NewReader(orgContent), "test.org")

	// Create a minimal org file struct
	type testOrgFile struct {
		title string
		doc   *goorg.Document
	}

	orgFile := &testOrgFile{
		title: "Test File",
		doc:   doc,
	}

	// Simulate what renderDocument does
	styles := NewStyles(r)
	renderer := NewRenderer(styles, 80)
	output := renderer.RenderNodes(orgFile.doc.Nodes)

	t.Log("=== Model renderDocument output ===")
	t.Log(output)
	t.Logf("\nRaw: %q", output)

	if !strings.Contains(output, "\x1b[1") {
		t.Error("Bold code missing from Model output")
	}
	if !strings.Contains(output, "\x1b[3") {
		t.Error("Italic code missing from Model output")
	}
}
