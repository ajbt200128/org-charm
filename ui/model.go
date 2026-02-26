package ui

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"org-charm/org"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

// View represents which view is currently active
type View int

const (
	ViewFileList View = iota
	ViewDocument
	ViewCredits
)

// Animation types
type AnimationType int

const (
	AnimNone AnimationType = iota
	AnimWaveRipple // Wave ripple on initial connection
	AnimPoof       // Poof/scatter effect on view toggle
)

// Animation constants
const (
	animFPS       = 60
	animFrequency = 2.5  // Lower = slower animation
	animDamping   = 1.0  // Critically damped for smooth motion without overshoot
)

// Particle characters for poof effect
var poofChars = []rune{'·', '∘', '°', '⋅', '✦', '✧', '∗', '⁕', '※', ' '}

// animTickMsg is sent on each animation frame
type animTickMsg time.Time

// secureRandInt returns a random int in [0, max) using crypto/rand
func secureRandInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

// secureRandRune returns a random rune from the given slice
func secureRandRune(runes []rune) rune {
	return runes[secureRandInt(len(runes))]
}

// Model is the bubbletea model for the org file viewer
type Model struct {
	styles   *Styles
	renderer *lipgloss.Renderer

	// Window dimensions
	width  int
	height int

	// File list state
	files         []string
	orgFiles      []*org.OrgFile
	selectedIndex int

	// Index file for main page (optional)
	indexFile *org.OrgFile

	// Current view
	currentView View

	// Viewport for scrolling document content
	viewport viewport.Model
	ready    bool

	// Current document being viewed
	currentDoc *org.OrgFile

	// Show help overlay
	showHelp bool

	// Show raw org content instead of rendered
	rawView bool

	// Changelog content for credits view
	changelog string

	// Animation state
	animType        AnimationType
	animSpring      harmonica.Spring
	animValue       float64 // Current animation progress (0.0 to 1.0)
	animVelocity    float64 // Current velocity for spring physics
	animTarget      float64 // Target value
	animFromContent string  // Content before transition (for poof)
	animToContent   string  // Content after transition (for poof)
	animContent     string  // Original content to reveal (for wave)
}

// NewModel creates a new Model with the given renderer and org files directory
func NewModel(renderer *lipgloss.Renderer, files []string, changelog string) Model {
	m := Model{
		renderer:      renderer,
		styles:        NewStyles(renderer),
		changelog:     changelog,
		files:         files,
		orgFiles:      make([]*org.OrgFile, 0),
		selectedIndex: 0,
		currentView:   ViewFileList,
		showHelp:      false,
		// Initialize animation - start with wave ripple
		animType:     AnimWaveRipple,
		animSpring:   harmonica.NewSpring(harmonica.FPS(animFPS), animFrequency, animDamping),
		animValue:    0.0,
		animVelocity: 0.0,
		animTarget:   1.0,
	}

	// Parse all org files, separating index.org
	for _, f := range files {
		if orgFile, err := org.ParseFile(f); err == nil {
			if strings.HasSuffix(strings.ToLower(f), "index.org") {
				m.indexFile = orgFile
			} else {
				m.orgFiles = append(m.orgFiles, orgFile)
			}
		}
	}

	return m
}

// animTick returns a command that sends animation tick messages
func animTick() tea.Cmd {
	return tea.Tick(time.Second/animFPS, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	// Start entrance animation
	return animTick()
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case animTickMsg:
		if m.animType != AnimNone {
			// Update spring physics
			m.animValue, m.animVelocity = m.animSpring.Update(m.animValue, m.animVelocity, m.animTarget)

			// Check if animation is complete (must be very close to target with low velocity)
			if m.animValue > 0.99 && abs(m.animVelocity) < 0.005 {
				m.animValue = 1.0
				m.animVelocity = 0.0
				m.animType = AnimNone
				m.animFromContent = ""
				m.animToContent = ""
			} else {
				// Continue animation
				cmds = append(cmds, animTick())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMargins)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = false
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMargins
		}

		if m.currentDoc != nil {
			m.viewport.SetContent(m.renderDocument(m.currentDoc))
		}

	case tea.KeyMsg:
		// Handle help toggle first
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// If help is shown, any key closes it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.currentView == ViewDocument {
				m.currentView = ViewFileList
				m.currentDoc = nil
				m.rawView = false
			} else if m.currentView == ViewCredits {
				m.currentView = ViewFileList
			}

		case "c":
			if m.currentView == ViewFileList {
				m.currentView = ViewCredits
				m.viewport.SetContent(m.renderCreditsContent())
				m.viewport.GotoTop()
			}

		case "up", "k":
			if m.currentView == ViewFileList {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}

		case "down", "j":
			if m.currentView == ViewFileList {
				if m.selectedIndex < len(m.orgFiles)-1 {
					m.selectedIndex++
				}
			}

		case "home", "g":
			if m.currentView == ViewFileList {
				m.selectedIndex = 0
			} else {
				m.viewport.GotoTop()
			}

		case "end", "G":
			if m.currentView == ViewFileList {
				m.selectedIndex = len(m.orgFiles) - 1
			} else {
				m.viewport.GotoBottom()
			}

		case "enter", "l", "right":
			if m.currentView == ViewFileList && len(m.orgFiles) > 0 {
				m.currentDoc = m.orgFiles[m.selectedIndex]
				m.currentView = ViewDocument
				m.viewport.SetContent(m.renderDocument(m.currentDoc))
				m.viewport.GotoTop()
			}

		case "h", "left":
			if m.currentView == ViewDocument {
				m.currentView = ViewFileList
				m.currentDoc = nil
				m.rawView = false
			}

		case "r":
			if m.currentView == ViewDocument && m.animType == AnimNone {
				// Capture current content for poof animation
				m.animFromContent = m.viewport.View()

				// Toggle view mode
				m.rawView = !m.rawView
				if m.rawView {
					m.viewport.SetContent(m.currentDoc.RawContent)
				} else {
					m.viewport.SetContent(m.renderDocument(m.currentDoc))
				}
				m.viewport.GotoTop()

				// Capture new content
				m.animToContent = m.viewport.View()

				// Start poof animation
				m.animType = AnimPoof
				m.animValue = 0.0
				m.animVelocity = 0.0
				m.animTarget = 1.0
				cmds = append(cmds, animTick())
			}

		case "n", "tab":
			// Next document
			if m.currentView == ViewDocument && len(m.orgFiles) > 1 {
				m.selectedIndex = (m.selectedIndex + 1) % len(m.orgFiles)
				m.currentDoc = m.orgFiles[m.selectedIndex]
				m.rawView = false
				m.viewport.SetContent(m.renderDocument(m.currentDoc))
				m.viewport.GotoTop()
			}

		case "p", "shift+tab":
			// Previous document
			if m.currentView == ViewDocument && len(m.orgFiles) > 1 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.orgFiles) - 1
				}
				m.currentDoc = m.orgFiles[m.selectedIndex]
				m.rawView = false
				m.viewport.SetContent(m.renderDocument(m.currentDoc))
				m.viewport.GotoTop()
			}
		}
	}

	// Handle viewport updates when viewing document or credits
	if (m.currentView == ViewDocument || m.currentView == ViewCredits) && !m.showHelp {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return m.styles.App.Render("Loading...")
	}

	var content string
	switch m.currentView {
	case ViewFileList:
		content = m.renderFileList()
	case ViewDocument:
		content = m.renderDocumentView()
	case ViewCredits:
		content = m.renderCreditsView()
	}

	// Overlay help if shown
	if m.showHelp {
		content = m.renderHelp()
	}

	// Apply wave animation (entrance only)
	if m.animType == AnimWaveRipple {
		content = m.applyWaveRipple(content)
	}
	// Note: Poof animation is applied within renderDocumentView

	return content
}

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// applyWaveRipple creates a radial wave effect that reveals content from the center
func (m Model) applyWaveRipple(content string) string {
	// When animation is nearly complete, return original content cleanly
	if m.animValue > 0.95 {
		return content
	}

	lines := strings.Split(content, "\n")
	centerY := m.height / 2
	centerX := m.width / 2

	// Maximum distance from center to corner
	maxDist := math.Sqrt(float64(centerX*centerX + centerY*centerY))

	// Current wave radius based on animation progress
	waveRadius := m.animValue * maxDist * 1.15
	waveWidth := maxDist * 0.12

	// Blue color codes for the wave
	blueLight := "\033[38;2;122;162;247m"
	blueMed := "\033[38;2;86;95;137m"
	blueDark := "\033[38;2;59;66;97m"
	reset := "\033[0m"

	var result strings.Builder

	for y, line := range lines {
		if y > 0 {
			result.WriteString("\n")
		}

		// Process the line, tracking visual column position
		visualCol := 0
		inEscape := false
		escapeSeq := strings.Builder{}

		for _, r := range line {
			// Handle ANSI escape sequences
			if r == '\033' {
				inEscape = true
				escapeSeq.Reset()
				escapeSeq.WriteRune(r)
				continue
			}
			if inEscape {
				escapeSeq.WriteRune(r)
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					inEscape = false
					// Calculate distance for this visual position
					dx := float64(visualCol - centerX)
					dy := float64(y - centerY)
					dist := math.Sqrt(dx*dx + dy*dy)
					// Only output escape sequence if inside wave (content visible)
					if dist < waveRadius-waveWidth {
						result.WriteString(escapeSeq.String())
					}
				}
				continue
			}

			// Regular character - calculate distance
			dx := float64(visualCol - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < waveRadius-waveWidth {
				// Inside the wave - show content (revealed)
				result.WriteRune(r)
			} else if dist < waveRadius {
				// On the wave crest - show blue wave character
				wavePos := (waveRadius - dist) / waveWidth
				if wavePos > 0.7 {
					result.WriteString(blueLight + "░" + reset)
				} else if wavePos > 0.4 {
					result.WriteString(blueMed + "▒" + reset)
				} else {
					result.WriteString(blueDark + "▓" + reset)
				}
			} else {
				// Outside the wave - dark/hidden
				result.WriteRune(' ')
			}

			visualCol++
		}

		// Pad to full width with wave effect
		for visualCol < m.width {
			dx := float64(visualCol - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < waveRadius-waveWidth {
				result.WriteRune(' ')
			} else if dist < waveRadius {
				wavePos := (waveRadius - dist) / waveWidth
				if wavePos > 0.7 {
					result.WriteString(blueLight + "░" + reset)
				} else if wavePos > 0.4 {
					result.WriteString(blueMed + "▒" + reset)
				} else {
					result.WriteString(blueDark + "▓" + reset)
				}
			} else {
				result.WriteRune(' ')
			}
			visualCol++
		}
	}

	return result.String()
}

// applyPoofToViewport applies poof animation to viewport content, transitioning from old to new
func (m Model) applyPoofToViewport(fromContent, toContent string) string {
	// When animation is nearly complete, return the target content cleanly
	if m.animValue > 0.95 {
		return toContent
	}

	fromLines := strings.Split(fromContent, "\n")
	toLines := strings.Split(toContent, "\n")

	// Strip ANSI for visual calculations
	fromClean := make([]string, len(fromLines))
	toClean := make([]string, len(toLines))
	for i, l := range fromLines {
		fromClean[i] = stripANSI(l)
	}
	for i, l := range toLines {
		toClean[i] = stripANSI(l)
	}

	// Ensure both have the same number of lines
	maxLines := len(fromLines)
	if len(toLines) > maxLines {
		maxLines = len(toLines)
	}

	centerX := m.width / 2
	centerY := maxLines / 2
	maxDist := math.Sqrt(float64(centerX*centerX + centerY*centerY))
	if maxDist < 1 {
		maxDist = 1
	}

	var result strings.Builder

	for y := 0; y < maxLines; y++ {
		if y > 0 {
			result.WriteString("\n")
		}

		// Get the source lines (or empty if beyond range)
		var fromLineClean, toLineClean string
		if y < len(fromClean) {
			fromLineClean = fromClean[y]
		}
		if y < len(toClean) {
			toLineClean = toClean[y]
		}

		// Determine max visual width
		maxCols := len([]rune(fromLineClean))
		if len([]rune(toLineClean)) > maxCols {
			maxCols = len([]rune(toLineClean))
		}
		if maxCols < m.width {
			maxCols = m.width
		}

		fromRunes := []rune(fromLineClean)
		toRunes := []rune(toLineClean)

		for x := 0; x < maxCols; x++ {
			// Get characters at this position
			var fromR, toR rune = ' ', ' '
			if x < len(fromRunes) {
				fromR = fromRunes[x]
			}
			if x < len(toRunes) {
				toR = toRunes[x]
			}

			// Position-based phase offset for organic ripple effect
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)
			distFactor := dist / maxDist * 0.2

			localAnim := m.animValue - distFactor
			if localAnim < 0 {
				localAnim = 0
			} else if localAnim > 1 {
				localAnim = 1
			}

			// Three phases: show old -> scatter -> show new
			if localAnim < 0.3 {
				// Phase 1: Show old content, starting to scatter
				scatterChance := localAnim / 0.3
				if secureRandInt(100) < int(scatterChance*70) {
					result.WriteRune(secureRandRune(poofChars))
				} else {
					result.WriteRune(fromR)
				}
			} else if localAnim < 0.7 {
				// Phase 2: Maximum scatter - particles
				if secureRandInt(100) < 75 {
					result.WriteRune(secureRandRune(poofChars))
				} else {
					result.WriteRune(' ')
				}
			} else {
				// Phase 3: Reform into new content
				reformProgress := (localAnim - 0.7) / 0.3
				if secureRandInt(100) < int(reformProgress*100) {
					result.WriteRune(toR)
				} else {
					result.WriteRune(secureRandRune(poofChars))
				}
			}
		}
	}

	return result.String()
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (m Model) renderFileList() string {
	var b strings.Builder

	// If we have an index.org, render it as the main page header
	if m.indexFile != nil {
		renderer := NewRenderer(m.styles, m.width-8)

		// Render index title if present
		if title := m.indexFile.Title(); title != "" {
			b.WriteString(m.styles.DocTitle.Width(m.width - 8).Render(title))
			b.WriteString("\n\n")
		}

		// Render index content
		b.WriteString(renderer.RenderNodes(m.indexFile.Document.Nodes))
		b.WriteString("\n")

		// Section header for file list
		b.WriteString(m.styles.Heading2.Render("★★ Documents"))
		b.WriteString("\n\n")
	} else {
		// Default header
		headerText := "  📚 Org Files"
		header := m.styles.Header.Width(m.width - 4).Render(headerText)
		b.WriteString(header)
		b.WriteString("\n\n")
	}

	// File list
	if len(m.orgFiles) == 0 {
		emptyMsg := m.styles.Paragraph.Render("No .org files found in the directory.")
		b.WriteString(emptyMsg)
	} else {
		// Calculate list area
		listWidth := m.width - 8

		for i, f := range m.orgFiles {
			title := f.Title()
			author := f.Author()
			date := f.Date()

			// Build metadata line
			var meta strings.Builder
			if author != "" {
				meta.WriteString(author)
			}
			if date != "" {
				if meta.Len() > 0 {
					meta.WriteString(" • ")
				}
				meta.WriteString(date)
			}

			// Render file item
			var line string
			if i == m.selectedIndex {
				// Selected item with arrow indicator
				indicator := "▸ "
				titleStyled := m.styles.FileItemActive.Width(listWidth - 2).Render(indicator + title)
				line = titleStyled
				if meta.Len() > 0 {
					line += "\n  " + m.styles.FileMeta.Render(meta.String())
				}
			} else {
				titleStyled := m.styles.FileItem.Render(title)
				line = titleStyled
			}

			b.WriteString(line)
			b.WriteString("\n")

			// Add spacing between items
			if i < len(m.orgFiles)-1 {
				b.WriteString("\n")
			}
		}
	}

	// Footer
	b.WriteString("\n")
	help := m.renderHelpBar([]helpItem{
		{"↑/↓", "navigate"},
		{"enter", "open"},
		{"c", "credits"},
		{"?", "help"},
		{"q", "quit"},
	})
	b.WriteString(help)

	return m.styles.App.Render(b.String())
}

func (m Model) renderCreditsView() string {
	var b strings.Builder

	// Header
	header := m.styles.Header.Width(m.width - 4).Render("  ✨ Credits & Changelog")
	b.WriteString(header)
	b.WriteString("\n")

	// Viewport content
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer
	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	scrollInfo := m.styles.StatusBar.Render(" " + scrollPercent + " ")
	help := m.renderHelpBar([]helpItem{
		{"↑/↓", "scroll"},
		{"esc", "back"},
		{"q", "quit"},
	})

	footer := lipgloss.JoinHorizontal(lipgloss.Center, scrollInfo, "  ", help)
	b.WriteString(footer)

	return m.styles.App.Render(b.String())
}

func (m Model) renderCreditsContent() string {
	var b strings.Builder

	// Authors section
	b.WriteString(m.styles.DocTitle.Width(m.width - 12).Render("Authors"))
	b.WriteString("\n\n")

	authors := []struct {
		name string
		role string
	}{
		{"Adam Tao", "Creator & Maintainer"},
		{"Claude (Anthropic)", "AI Pair Programmer"},
	}

	for _, author := range authors {
		b.WriteString(m.styles.Bold.Render("  • " + author.name))
		b.WriteString(m.styles.HelpText.Render(" — " + author.role))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.HRule.Width(m.width - 12).Render(""))
	b.WriteString("\n\n")

	// Changelog section
	b.WriteString(m.styles.DocTitle.Width(m.width - 12).Render("Changelog"))
	b.WriteString("\n\n")

	// Parse and render the changelog with simple formatting
	lines := strings.Split(m.changelog, "\n")
	for _, line := range lines {
		// Skip the main title
		if strings.HasPrefix(line, "# ") {
			continue
		}

		// Format headers
		if strings.HasPrefix(line, "## ") {
			version := strings.TrimPrefix(line, "## ")
			b.WriteString(m.styles.Heading2.Render(version))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "### ") {
			section := strings.TrimPrefix(line, "### ")
			b.WriteString(m.styles.Heading3.Render("  " + section))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "- ") {
			item := strings.TrimPrefix(line, "- ")
			b.WriteString(m.styles.Paragraph.Render("    • " + item))
			b.WriteString("\n")
		} else if strings.HasPrefix(line, "  - ") {
			item := strings.TrimPrefix(line, "  - ")
			b.WriteString(m.styles.HelpText.Render("      ◦ " + item))
			b.WriteString("\n")
		} else if strings.TrimSpace(line) != "" {
			b.WriteString(m.styles.Paragraph.Render("  " + line))
			b.WriteString("\n")
		} else {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderDocumentView() string {
	var b strings.Builder

	// Header with document info
	title := m.currentDoc.Title()
	author := m.currentDoc.Author()
	date := m.currentDoc.Date()

	headerContent := "  📄 " + title
	if author != "" {
		headerContent += " — " + author
	}
	if date != "" {
		headerContent += " (" + date + ")"
	}

	header := m.styles.Header.Width(m.width - 4).Render(headerContent)
	b.WriteString(header)
	b.WriteString("\n")

	// Viewport content - apply poof animation if active
	viewportContent := m.viewport.View()
	if m.animType == AnimPoof {
		viewportContent = m.applyPoofToViewport(m.animFromContent, m.animToContent)
	}
	b.WriteString(viewportContent)
	b.WriteString("\n")

	// Footer with scroll info and help
	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	scrollInfo := m.styles.StatusBar.Render(" " + scrollPercent + " ")

	var rawToggle string
	if m.rawView {
		rawToggle = "rendered"
	} else {
		rawToggle = "raw"
	}
	help := m.renderHelpBar([]helpItem{
		{"↑/↓", "scroll"},
		{"n/p", "next/prev"},
		{"r", rawToggle},
		{"esc", "back"},
		{"q", "quit"},
	})

	footer := lipgloss.JoinHorizontal(lipgloss.Center, scrollInfo, "  ", help)
	b.WriteString(footer)

	return m.styles.App.Render(b.String())
}

func (m Model) renderDocument(doc *org.OrgFile) string {
	var b strings.Builder
	renderer := NewRenderer(m.styles, m.width-8)

	// Render document metadata header
	title := doc.Title()
	author := doc.Author()
	date := doc.Date()

	if title != "" || author != "" || date != "" {
		// Title
		if title != "" {
			b.WriteString(m.styles.DocTitle.Width(m.width - 12).Render(title))
			b.WriteString("\n")
		}

		// Author and date line
		var meta []string
		if author != "" {
			meta = append(meta, m.styles.DocAuthor.Render("by "+author))
		}
		if date != "" {
			meta = append(meta, m.styles.DocDate.Render(date))
		}
		if len(meta) > 0 {
			b.WriteString(strings.Join(meta, m.styles.HelpText.Render(" • ")))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Render document content
	b.WriteString(renderer.RenderNodes(doc.Document.Nodes))
	return b.String()
}

type helpItem struct {
	key  string
	desc string
}

func (m Model) renderHelpBar(items []helpItem) string {
	var parts []string
	for _, item := range items {
		part := m.styles.HelpKey.Render(item.key) + " " + m.styles.HelpText.Render(item.desc)
		parts = append(parts, part)
	}
	return strings.Join(parts, m.styles.HelpText.Render(" • "))
}

func (m Model) renderHelp() string {
	var b strings.Builder

	title := m.styles.DocTitle.Width(m.width - 8).Render("  ⌨️  Keyboard Shortcuts")
	b.WriteString(title)
	b.WriteString("\n\n")

	sections := []struct {
		name  string
		items []helpItem
	}{
		{
			name: "Navigation",
			items: []helpItem{
				{"↑ / k", "Move up"},
				{"↓ / j", "Move down"},
				{"← / h", "Go back"},
				{"→ / l / Enter", "Open / Select"},
				{"g / Home", "Go to top"},
				{"G / End", "Go to bottom"},
			},
		},
		{
			name: "Document View",
			items: []helpItem{
				{"Page Up / Ctrl+u", "Scroll up"},
				{"Page Down / Ctrl+d", "Scroll down"},
				{"n / Tab", "Next document"},
				{"p / Shift+Tab", "Previous document"},
				{"r", "Toggle raw/rendered view"},
				{"Esc", "Return to file list"},
			},
		},
		{
			name: "General",
			items: []helpItem{
				{"c", "Show credits & changelog"},
				{"?", "Toggle this help"},
				{"q / Ctrl+c", "Quit"},
			},
		},
	}

	for _, section := range sections {
		sectionTitle := m.styles.Heading3.Render("  " + section.name)
		b.WriteString(sectionTitle)
		b.WriteString("\n")

		for _, item := range section.items {
			line := "    " + m.styles.HelpKey.Render(fmt.Sprintf("%-20s", item.key)) +
				m.styles.HelpText.Render(item.desc)
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.HelpText.Render("  Press any key to close this help"))

	return m.styles.App.Render(b.String())
}
