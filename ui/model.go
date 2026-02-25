package ui

import (
	"fmt"
	"strings"

	"org-charm/org"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents which view is currently active
type View int

const (
	ViewFileList View = iota
	ViewDocument
)

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
}

// NewModel creates a new Model with the given renderer and org files directory
func NewModel(renderer *lipgloss.Renderer, files []string) Model {
	m := Model{
		renderer:      renderer,
		styles:        NewStyles(renderer),
		files:         files,
		orgFiles:      make([]*org.OrgFile, 0),
		selectedIndex: 0,
		currentView:   ViewFileList,
		showHelp:      false,
	}

	// Parse all org files
	for _, f := range files {
		if orgFile, err := org.ParseFile(f); err == nil {
			m.orgFiles = append(m.orgFiles, orgFile)
		}
	}

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
			if m.currentView == ViewDocument {
				m.rawView = !m.rawView
				if m.rawView {
					m.viewport.SetContent(m.currentDoc.RawContent)
				} else {
					m.viewport.SetContent(m.renderDocument(m.currentDoc))
				}
				m.viewport.GotoTop()
			}
		}
	}

	// Handle viewport updates when viewing document
	if m.currentView == ViewDocument && !m.showHelp {
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
	}

	// Overlay help if shown
	if m.showHelp {
		content = m.renderHelp()
	}

	return content
}

func (m Model) renderFileList() string {
	var b strings.Builder

	// Header
	headerText := "  📚 Org Files"
	header := m.styles.Header.Width(m.width - 4).Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

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
		{"?", "help"},
		{"q", "quit"},
	})
	b.WriteString(help)

	return m.styles.App.Render(b.String())
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

	// Viewport content
	b.WriteString(m.viewport.View())
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
		{"r", rawToggle},
		{"esc", "back"},
		{"?", "help"},
		{"q", "quit"},
	})

	footer := lipgloss.JoinHorizontal(lipgloss.Center, scrollInfo, "  ", help)
	b.WriteString(footer)

	return m.styles.App.Render(b.String())
}

func (m Model) renderDocument(doc *org.OrgFile) string {
	renderer := NewRenderer(m.styles, m.width-8)
	return renderer.RenderNodes(doc.Document.Nodes)
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
				{"r", "Toggle raw/rendered view"},
				{"Esc", "Return to file list"},
			},
		},
		{
			name: "General",
			items: []helpItem{
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
