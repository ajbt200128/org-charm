package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all the lipgloss styles for the UI
type Styles struct {
	// App frame
	App       lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
	StatusBar lipgloss.Style

	// File list
	FileList         lipgloss.Style
	FileItem         lipgloss.Style
	FileItemSelected lipgloss.Style
	FileItemActive   lipgloss.Style
	FileDir          lipgloss.Style
	FileMeta         lipgloss.Style

	// Document metadata
	DocTitle  lipgloss.Style
	DocAuthor lipgloss.Style
	DocDate   lipgloss.Style

	// Headings
	Heading1 lipgloss.Style
	Heading2 lipgloss.Style
	Heading3 lipgloss.Style
	Heading4 lipgloss.Style

	// TODO/DONE states
	Todo     lipgloss.Style
	Done     lipgloss.Style
	Priority lipgloss.Style
	Tag      lipgloss.Style

	// Text content
	Paragraph lipgloss.Style

	// Lists
	ListBullet      lipgloss.Style
	ListItem        lipgloss.Style
	DescTerm        lipgloss.Style
	DescSeparator   lipgloss.Style
	CheckboxEmpty   lipgloss.Style
	CheckboxDone    lipgloss.Style
	CheckboxPartial lipgloss.Style

	// Code blocks
	BlockHeader lipgloss.Style
	CodeBlock   lipgloss.Style
	Example     lipgloss.Style

	// Quotes and verse
	Quote  lipgloss.Style
	Verse  lipgloss.Style
	Center lipgloss.Style

	// Tables
	TableBorder lipgloss.Style
	TableHeader lipgloss.Style
	TableCell   lipgloss.Style

	// Inline formatting
	Bold          lipgloss.Style
	Italic        lipgloss.Style
	Underline     lipgloss.Style
	Strikethrough lipgloss.Style
	Verbatim      lipgloss.Style
	InlineCode    lipgloss.Style
	Link          lipgloss.Style

	// Other elements
	HRule           lipgloss.Style
	Keyword         lipgloss.Style
	KeywordValue    lipgloss.Style
	DrawerHeader    lipgloss.Style
	Property        lipgloss.Style
	Timestamp       lipgloss.Style
	Footnote           lipgloss.Style
	FootnoteLabel      lipgloss.Style
	FootnoteContent    lipgloss.Style
	FootnoteRef        lipgloss.Style
	FootnoteNestedLabel1 lipgloss.Style // a., b., c.
	FootnoteNestedLabel2 lipgloss.Style // i., ii., iii.
	FootnoteNestedLabel3 lipgloss.Style // α, β, γ
	FootnoteNestedRef1   lipgloss.Style
	FootnoteNestedRef2   lipgloss.Style
	FootnoteNestedRef3   lipgloss.Style
	Statistics         lipgloss.Style

	// Planning keywords
	Scheduled lipgloss.Style
	Deadline  lipgloss.Style
	Closed    lipgloss.Style

	// Help/hints
	HelpKey  lipgloss.Style
	HelpText lipgloss.Style
}

// Colors - a cohesive palette
var (
	// Base colors
	colorBg        = lipgloss.Color("#1a1b26")
	colorFg        = lipgloss.Color("#c0caf5")
	colorSubtle    = lipgloss.Color("#565f89")
	colorHighlight = lipgloss.Color("#7aa2f7")
	colorAccent    = lipgloss.Color("#bb9af7")

	// Semantic colors
	colorRed     = lipgloss.Color("#f7768e")
	colorGreen   = lipgloss.Color("#9ece6a")
	colorYellow  = lipgloss.Color("#e0af68")
	colorBlue    = lipgloss.Color("#7aa2f7")
	colorMagenta = lipgloss.Color("#bb9af7")
	colorCyan    = lipgloss.Color("#7dcfff")
	colorOrange  = lipgloss.Color("#ff9e64")

	// Heading colors (rainbow progression)
	colorH1 = lipgloss.Color("#f7768e") // Red
	colorH2 = lipgloss.Color("#ff9e64") // Orange
	colorH3 = lipgloss.Color("#e0af68") // Yellow
	colorH4 = lipgloss.Color("#9ece6a") // Green
)

// NewStyles creates a new Styles instance with the given renderer
func NewStyles(r *lipgloss.Renderer) *Styles {
	s := &Styles{}

	// ═══════════════════════════════════════════════════════════════════
	// App Frame
	// ═══════════════════════════════════════════════════════════════════

	s.App = r.NewStyle().
		Padding(1, 2)

	s.Header = r.NewStyle().
		Bold(true).
		Foreground(colorHighlight).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 2).
		MarginBottom(1)

	s.Footer = r.NewStyle().
		Foreground(colorSubtle).
		Padding(0, 1).
		MarginTop(1)

	s.StatusBar = r.NewStyle().
		Foreground(colorFg).
		Background(colorHighlight).
		Padding(0, 1)

	// ═══════════════════════════════════════════════════════════════════
	// File List
	// ═══════════════════════════════════════════════════════════════════

	s.FileList = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSubtle).
		Padding(1, 2)

	s.FileItem = r.NewStyle().
		Foreground(colorFg).
		PaddingLeft(2)

	s.FileItemSelected = r.NewStyle().
		Foreground(colorHighlight).
		Bold(true).
		PaddingLeft(0)

	s.FileItemActive = r.NewStyle().
		Foreground(colorAccent).
		Background(lipgloss.Color("#24283b")).
		Bold(true).
		PaddingLeft(0).
		PaddingRight(2)

	s.FileDir = r.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		PaddingLeft(2)

	s.FileMeta = r.NewStyle().
		Foreground(colorSubtle).
		Italic(true)

	// ═══════════════════════════════════════════════════════════════════
	// Document Metadata
	// ═══════════════════════════════════════════════════════════════════

	s.DocTitle = r.NewStyle().
		Bold(true).
		Foreground(colorHighlight).
		MarginBottom(1).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderBottom(true).
		BorderForeground(colorSubtle).
		Padding(0, 1)

	s.DocAuthor = r.NewStyle().
		Foreground(colorCyan).
		Italic(true)

	s.DocDate = r.NewStyle().
		Foreground(colorSubtle)

	// ═══════════════════════════════════════════════════════════════════
	// Headings
	// ═══════════════════════════════════════════════════════════════════

	s.Heading1 = r.NewStyle().
		Bold(true).
		Foreground(colorH1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorH1)

	s.Heading2 = r.NewStyle().
		Bold(true).
		Foreground(colorH2)

	s.Heading3 = r.NewStyle().
		Bold(true).
		Foreground(colorH3)

	s.Heading4 = r.NewStyle().
		Foreground(colorH4)

	// ═══════════════════════════════════════════════════════════════════
	// TODO States
	// ═══════════════════════════════════════════════════════════════════

	s.Todo = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1a1b26")).
		Background(colorRed).
		Padding(0, 1)

	s.Done = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1a1b26")).
		Background(colorGreen).
		Padding(0, 1)

	s.Priority = r.NewStyle().
		Bold(true).
		Foreground(colorOrange)

	s.Tag = r.NewStyle().
		Foreground(colorMagenta).
		Italic(true)

	// ═══════════════════════════════════════════════════════════════════
	// Text Content
	// ═══════════════════════════════════════════════════════════════════

	s.Paragraph = r.NewStyle().
		Foreground(colorFg)

	// ═══════════════════════════════════════════════════════════════════
	// Lists
	// ═══════════════════════════════════════════════════════════════════

	s.ListBullet = r.NewStyle().
		Foreground(colorCyan).
		Bold(true)

	s.ListItem = r.NewStyle().
		Foreground(colorFg)

	s.DescTerm = r.NewStyle().
		Bold(true).
		Foreground(colorYellow)

	s.DescSeparator = r.NewStyle().
		Foreground(colorSubtle).
		Bold(true)

	s.CheckboxEmpty = r.NewStyle().
		Foreground(colorSubtle)

	s.CheckboxDone = r.NewStyle().
		Foreground(colorGreen)

	s.CheckboxPartial = r.NewStyle().
		Foreground(colorYellow)

	// ═══════════════════════════════════════════════════════════════════
	// Code Blocks
	// ═══════════════════════════════════════════════════════════════════

	s.BlockHeader = r.NewStyle().
		Foreground(colorSubtle)

	s.CodeBlock = r.NewStyle().
		Background(lipgloss.Color("#1f2335")).
		Foreground(colorFg).
		Padding(1, 2).
		MarginTop(0).
		MarginBottom(0)

	s.Example = r.NewStyle().
		Background(lipgloss.Color("#1f2335")).
		Foreground(colorCyan).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	// ═══════════════════════════════════════════════════════════════════
	// Quotes and Verse
	// ═══════════════════════════════════════════════════════════════════

	s.Quote = r.NewStyle().
		Foreground(colorAccent).
		Italic(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(colorAccent).
		PaddingLeft(2).
		MarginTop(1).
		MarginBottom(1)

	s.Verse = r.NewStyle().
		Foreground(colorCyan).
		Italic(true).
		PaddingLeft(4).
		MarginTop(1).
		MarginBottom(1)

	s.Center = r.NewStyle().
		Foreground(colorFg).
		Align(lipgloss.Center)

	// ═══════════════════════════════════════════════════════════════════
	// Tables
	// ═══════════════════════════════════════════════════════════════════

	s.TableBorder = r.NewStyle().
		Foreground(colorSubtle)

	s.TableHeader = r.NewStyle().
		Bold(true).
		Foreground(colorHighlight).
		Background(lipgloss.Color("#24283b"))

	s.TableCell = r.NewStyle().
		Foreground(colorFg)

	// ═══════════════════════════════════════════════════════════════════
	// Inline Formatting - distinct colors for visibility
	// ═══════════════════════════════════════════════════════════════════

	s.Bold = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")) // White for bold

	s.Italic = r.NewStyle().
		Italic(true).
		Foreground(colorCyan) // Cyan for italic

	s.Underline = r.NewStyle().
		Underline(true).
		Foreground(colorYellow) // Yellow for underline

	s.Strikethrough = r.NewStyle().
		Strikethrough(true).
		Foreground(colorSubtle)

	s.Verbatim = r.NewStyle().
		Foreground(colorGreen).
		Background(lipgloss.Color("#1f2335"))

	s.InlineCode = r.NewStyle().
		Background(lipgloss.Color("#24283b")).
		Foreground(colorOrange)

	s.Link = r.NewStyle().
		Foreground(colorBlue).
		Underline(true)

	// ═══════════════════════════════════════════════════════════════════
	// Other Elements
	// ═══════════════════════════════════════════════════════════════════

	s.HRule = r.NewStyle().
		Foreground(colorSubtle)

	s.Keyword = r.NewStyle().
		Foreground(colorMagenta)

	s.KeywordValue = r.NewStyle().
		Foreground(colorFg)

	s.DrawerHeader = r.NewStyle().
		Foreground(colorSubtle).
		Italic(true)

	s.Property = r.NewStyle().
		Foreground(colorSubtle)

	s.Timestamp = r.NewStyle().
		Foreground(colorCyan).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1)

	s.Footnote = r.NewStyle().
		Foreground(colorYellow)

	s.FootnoteLabel = r.NewStyle().
		Bold(true).
		Foreground(colorYellow).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1)

	s.FootnoteContent = r.NewStyle().
		Foreground(colorFg).
		Italic(true)

	s.FootnoteRef = r.NewStyle().
		Foreground(colorYellow).
		Bold(true)

	// Nested footnote styles (level 1: a., b., c.)
	s.FootnoteNestedLabel1 = r.NewStyle().
		Bold(true).
		Foreground(colorCyan).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1)

	s.FootnoteNestedRef1 = r.NewStyle().
		Foreground(colorCyan).
		Bold(true)

	// Nested footnote styles (level 2: i., ii., iii.)
	s.FootnoteNestedLabel2 = r.NewStyle().
		Bold(true).
		Foreground(colorMagenta).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1)

	s.FootnoteNestedRef2 = r.NewStyle().
		Foreground(colorMagenta).
		Bold(true)

	// Nested footnote styles (level 3: α, β, γ)
	s.FootnoteNestedLabel3 = r.NewStyle().
		Bold(true).
		Foreground(colorOrange).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1)

	s.FootnoteNestedRef3 = r.NewStyle().
		Foreground(colorOrange).
		Bold(true)

	s.Statistics = r.NewStyle().
		Foreground(colorGreen).
		Bold(true)

	// ═══════════════════════════════════════════════════════════════════
	// Planning Keywords
	// ═══════════════════════════════════════════════════════════════════

	s.Scheduled = r.NewStyle().
		Foreground(colorGreen).
		Bold(true)

	s.Deadline = r.NewStyle().
		Foreground(colorRed).
		Bold(true)

	s.Closed = r.NewStyle().
		Foreground(colorSubtle).
		Italic(true)

	// ═══════════════════════════════════════════════════════════════════
	// Help
	// ═══════════════════════════════════════════════════════════════════

	s.HelpKey = r.NewStyle().
		Foreground(colorHighlight).
		Bold(true)

	s.HelpText = r.NewStyle().
		Foreground(colorSubtle)

	return s
}
