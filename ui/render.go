package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	goorg "github.com/niklasfasching/go-org/org"
)

// Renderer handles rendering org document nodes to styled strings
type Renderer struct {
	styles *Styles
	width  int
}

// NewRenderer creates a new Renderer
func NewRenderer(styles *Styles, width int) *Renderer {
	return &Renderer{
		styles: styles,
		width:  width,
	}
}

// RenderNodes renders a slice of org nodes
func (r *Renderer) RenderNodes(nodes []goorg.Node) string {
	var b strings.Builder
	for _, node := range nodes {
		rendered := r.RenderNode(node)
		if rendered != "" {
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// RenderNode renders a single org node
func (r *Renderer) RenderNode(node goorg.Node) string {
	switch n := node.(type) {
	case goorg.Headline:
		return r.renderHeadline(n)
	case goorg.Block:
		return r.renderBlock(n)
	case goorg.Paragraph:
		return r.renderParagraph(n)
	case goorg.List:
		return r.renderList(n)
	case goorg.ListItem:
		return r.renderListItem(n, 0)
	case goorg.DescriptiveListItem:
		return r.renderDescriptiveListItem(n)
	case goorg.Table:
		return r.renderTable(n)
	case goorg.HorizontalRule:
		return r.renderHorizontalRule()
	case goorg.Keyword:
		return r.renderKeyword(n)
	case goorg.PropertyDrawer:
		return r.renderPropertyDrawer(n)
	case goorg.Drawer:
		return r.renderDrawer(n)
	case goorg.Example:
		return r.renderExample(n)
	case goorg.FootnoteDefinition:
		return r.renderFootnoteDefinition(n)
	default:
		return ""
	}
}

func (r *Renderer) renderHeadline(h goorg.Headline) string {
	var b strings.Builder

	// Build the headline text
	stars := strings.Repeat("★", h.Lvl)
	title := r.renderInlineNodes(h.Title)

	// Add TODO/DONE status with styling
	var status string
	if h.Status != "" {
		if h.Status == "DONE" {
			status = r.styles.Done.Render(h.Status) + " "
		} else {
			status = r.styles.Todo.Render(h.Status) + " "
		}
	}

	// Add priority
	var priority string
	if h.Priority != "" {
		priority = r.styles.Priority.Render("[#"+h.Priority+"]") + " "
	}

	// Add tags
	var tags string
	if len(h.Tags) > 0 {
		tags = " " + r.styles.Tag.Render(":"+strings.Join(h.Tags, ":")+":")
	}

	headline := fmt.Sprintf("%s %s%s%s%s", stars, status, priority, title, tags)

	// Style based on level
	var style lipgloss.Style
	switch h.Lvl {
	case 1:
		style = r.styles.Heading1
	case 2:
		style = r.styles.Heading2
	case 3:
		style = r.styles.Heading3
	default:
		style = r.styles.Heading4
	}

	b.WriteString(style.Render(headline))
	b.WriteString("\n")

	// Render children
	for _, child := range h.Children {
		rendered := r.RenderNode(child)
		if rendered != "" {
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (r *Renderer) renderBlock(block goorg.Block) string {
	name := strings.ToUpper(block.Name)

	switch name {
	case "SRC":
		return r.renderSourceBlock(block)
	case "QUOTE":
		return r.renderQuoteBlock(block)
	case "EXAMPLE":
		return r.renderExampleBlock(block)
	case "VERSE":
		return r.renderVerseBlock(block)
	case "CENTER":
		return r.renderCenterBlock(block)
	default:
		// Generic block
		content := r.extractBlockText(block.Children)
		return r.styles.CodeBlock.Width(r.width - 6).Render(content)
	}
}

func (r *Renderer) renderSourceBlock(block goorg.Block) string {
	content := r.extractBlockText(block.Children)
	lang := ""

	// Get language from parameters - first parameter is typically the language
	if len(block.Parameters) > 0 {
		lang = block.Parameters[0]
	}

	// Try to syntax highlight with chroma
	highlighted := r.highlightCode(content, lang)

	// Add language label
	headerWidth := r.width - 8
	if headerWidth < 10 {
		headerWidth = 10
	}

	var header string
	if lang != "" {
		langLabel := " " + lang + " "
		lineLen := headerWidth - len(langLabel) - 2
		if lineLen < 0 {
			lineLen = 0
		}
		header = r.styles.BlockHeader.Render("┌─" + langLabel + strings.Repeat("─", lineLen) + "┐")
	} else {
		header = r.styles.BlockHeader.Render("┌" + strings.Repeat("─", headerWidth) + "┐")
	}

	footer := r.styles.BlockHeader.Render("└" + strings.Repeat("─", headerWidth) + "┘")

	codeBlock := r.styles.CodeBlock.Width(r.width - 6).Render(highlighted)

	return header + "\n" + codeBlock + "\n" + footer
}

func (r *Renderer) highlightCode(code, lang string) string {
	if lang == "" {
		return code
	}

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Use a terminal-friendly style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	return buf.String()
}

func (r *Renderer) renderQuoteBlock(block goorg.Block) string {
	content := r.renderInlineNodes(block.Children)
	return r.styles.Quote.Width(r.width - 8).Render(content)
}

func (r *Renderer) renderExampleBlock(block goorg.Block) string {
	content := r.extractBlockText(block.Children)
	return r.styles.Example.Width(r.width - 6).Render(content)
}

func (r *Renderer) renderVerseBlock(block goorg.Block) string {
	content := r.extractBlockText(block.Children)
	return r.styles.Verse.Width(r.width - 6).Render(content)
}

func (r *Renderer) renderCenterBlock(block goorg.Block) string {
	content := r.renderInlineNodes(block.Children)
	return r.styles.Center.Width(r.width - 6).Render(content)
}

func (r *Renderer) renderParagraph(p goorg.Paragraph) string {
	content := r.renderInlineNodes(p.Children)
	return r.styles.Paragraph.Width(r.width - 4).Render(content)
}

func (r *Renderer) renderList(list goorg.List) string {
	var b strings.Builder

	for _, item := range list.Items {
		switch n := item.(type) {
		case goorg.ListItem:
			b.WriteString(r.renderListItem(n, 0))
			b.WriteString("\n")
		case goorg.DescriptiveListItem:
			b.WriteString(r.renderDescriptiveListItem(n))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (r *Renderer) renderListItem(item goorg.ListItem, indent int) string {
	var b strings.Builder

	indentStr := strings.Repeat("  ", indent)

	// Determine bullet style
	bullet := "•"
	if strings.HasPrefix(item.Bullet, "1") || strings.ContainsAny(item.Bullet, "0123456789") {
		bullet = item.Bullet
	}

	// Checkbox status
	var checkbox string
	switch item.Status {
	case "X", "x":
		checkbox = r.styles.CheckboxDone.Render("[✓]") + " "
	case "-":
		checkbox = r.styles.CheckboxPartial.Render("[~]") + " "
	case " ":
		checkbox = r.styles.CheckboxEmpty.Render("[ ]") + " "
	}

	// ListItem.Children contains block elements (usually Paragraph, but also nested List)
	// We need to extract and render the inline content from Paragraphs,
	// and recursively render nested Lists
	var content string
	var nestedContent string
	for _, child := range item.Children {
		switch c := child.(type) {
		case goorg.Paragraph:
			content += r.renderInlineNodes(c.Children)
		case goorg.List:
			// Nested list - render with increased indent
			nestedContent += "\n" + r.renderListWithIndent(c, indent+1)
		default:
			// For other block types, render them normally
			content += r.RenderNode(child)
		}
	}

	b.WriteString(indentStr)
	b.WriteString(r.styles.ListBullet.Render(bullet))
	b.WriteString(" ")
	b.WriteString(checkbox)
	b.WriteString(r.styles.ListItem.Render(content))
	b.WriteString(nestedContent)

	return b.String()
}

func (r *Renderer) renderListWithIndent(list goorg.List, indent int) string {
	var b strings.Builder

	for _, item := range list.Items {
		switch n := item.(type) {
		case goorg.ListItem:
			b.WriteString(r.renderListItem(n, indent))
			b.WriteString("\n")
		case goorg.DescriptiveListItem:
			// Descriptive list items with indent
			indentStr := strings.Repeat("  ", indent)
			term := r.renderInlineNodes(n.Term)
			details := r.renderInlineNodes(n.Details)
			b.WriteString(indentStr)
			b.WriteString(r.styles.ListBullet.Render("•") + " ")
			b.WriteString(r.styles.DescTerm.Render(term) + " ")
			b.WriteString(r.styles.DescSeparator.Render("::") + " ")
			b.WriteString(r.styles.ListItem.Render(details))
			b.WriteString("\n")
		}
	}

	return strings.TrimSuffix(b.String(), "\n")
}

func (r *Renderer) renderDescriptiveListItem(item goorg.DescriptiveListItem) string {
	term := r.renderInlineNodes(item.Term)
	details := r.renderInlineNodes(item.Details)

	return r.styles.ListBullet.Render("•") + " " +
		r.styles.DescTerm.Render(term) + " " +
		r.styles.DescSeparator.Render("::") + " " +
		r.styles.ListItem.Render(details)
}

func (r *Renderer) renderTable(table goorg.Table) string {
	var b strings.Builder

	if len(table.Rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, 0)
	for _, row := range table.Rows {
		if row.IsSpecial {
			continue
		}
		for i, col := range row.Columns {
			content := r.renderInlineNodes(col.Children)
			width := len(content)
			if width < 3 {
				width = 3
			}
			if i >= len(colWidths) {
				colWidths = append(colWidths, width)
			} else if width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	if len(colWidths) == 0 {
		return ""
	}

	// Helper to render a horizontal border
	renderBorder := func(left, mid, right, fill string) string {
		var sb strings.Builder
		sb.WriteString(r.styles.TableBorder.Render(left))
		for i, w := range colWidths {
			sb.WriteString(r.styles.TableBorder.Render(strings.Repeat(fill, w+2)))
			if i < len(colWidths)-1 {
				sb.WriteString(r.styles.TableBorder.Render(mid))
			}
		}
		sb.WriteString(r.styles.TableBorder.Render(right))
		return sb.String()
	}

	// Top border
	b.WriteString(renderBorder("╭", "┬", "╮", "─"))
	b.WriteString("\n")

	// Render rows
	for rowIdx, row := range table.Rows {
		if row.IsSpecial {
			// Separator row
			b.WriteString(renderBorder("├", "┼", "┤", "─"))
			b.WriteString("\n")
			continue
		}

		// Header row detection (first row before separator)
		isHeader := rowIdx == 0 && len(table.Rows) > 1 && table.Rows[1].IsSpecial

		var rowStr strings.Builder
		rowStr.WriteString(r.styles.TableBorder.Render("│"))
		for i, col := range row.Columns {
			content := r.renderInlineNodes(col.Children)
			width := 3
			if i < len(colWidths) {
				width = colWidths[i]
			}
			padded := fmt.Sprintf(" %-*s ", width, content)
			if isHeader {
				rowStr.WriteString(r.styles.TableHeader.Render(padded))
			} else {
				rowStr.WriteString(r.styles.TableCell.Render(padded))
			}
			rowStr.WriteString(r.styles.TableBorder.Render("│"))
		}
		b.WriteString(rowStr.String())
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(renderBorder("╰", "┴", "╯", "─"))

	return b.String()
}

func (r *Renderer) renderHorizontalRule() string {
	return r.styles.HRule.Render(strings.Repeat("─", r.width-4))
}

func (r *Renderer) renderKeyword(kw goorg.Keyword) string {
	// Skip rendering most keywords, but show some
	switch strings.ToUpper(kw.Key) {
	case "TITLE", "AUTHOR", "DATE", "OPTIONS":
		return "" // These are metadata, don't render
	default:
		return r.styles.Keyword.Render("#+"+kw.Key+": ") + r.styles.KeywordValue.Render(kw.Value)
	}
}

func (r *Renderer) renderPropertyDrawer(pd goorg.PropertyDrawer) string {
	var b strings.Builder
	b.WriteString(r.styles.DrawerHeader.Render(":PROPERTIES:"))
	b.WriteString("\n")
	for _, prop := range pd.Properties {
		if len(prop) >= 2 {
			b.WriteString(r.styles.Property.Render(fmt.Sprintf(":%s: %s", prop[0], prop[1])))
			b.WriteString("\n")
		}
	}
	b.WriteString(r.styles.DrawerHeader.Render(":END:"))
	return b.String()
}

func (r *Renderer) renderDrawer(d goorg.Drawer) string {
	var b strings.Builder
	b.WriteString(r.styles.DrawerHeader.Render(":" + d.Name + ":"))
	b.WriteString("\n")
	b.WriteString(r.renderInlineNodes(d.Children))
	b.WriteString("\n")
	b.WriteString(r.styles.DrawerHeader.Render(":END:"))
	return b.String()
}

func (r *Renderer) renderExample(ex goorg.Example) string {
	content := r.extractBlockText(ex.Children)
	return r.styles.Example.Width(r.width - 6).Render(content)
}

func (r *Renderer) renderFootnoteDefinition(fn goorg.FootnoteDefinition) string {
	content := r.renderInlineNodes(fn.Children)
	// Render footnote with a nice box
	label := r.styles.FootnoteLabel.Render("[" + fn.Name + "]")
	return label + " " + r.styles.FootnoteContent.Render(content)
}

// renderInlineNodes renders inline content (text, emphasis, links, etc.)
func (r *Renderer) renderInlineNodes(nodes []goorg.Node) string {
	var b strings.Builder
	for _, node := range nodes {
		b.WriteString(r.renderInlineNode(node))
	}
	return b.String()
}

func (r *Renderer) renderInlineNode(node goorg.Node) string {
	switch n := node.(type) {
	case goorg.Text:
		return r.renderText(n.Content)
	case goorg.Emphasis:
		return r.renderEmphasis(n)
	case goorg.RegularLink:
		return r.renderLink(n)
	case goorg.StatisticToken:
		return r.styles.Statistics.Render("[" + n.Content + "]")
	case goorg.Timestamp:
		return r.renderTimestamp(n)
	case goorg.FootnoteLink:
		return r.renderFootnoteLink(n)
	case goorg.ExplicitLineBreak:
		return "\n"
	case goorg.LineBreak:
		return "\n"
	default:
		// For unknown types, try to get string representation
		return fmt.Sprintf("%v", n)
	}
}

// renderText handles plain text with planning keyword detection and inactive timestamps
func (r *Renderer) renderText(content string) string {
	// Check for planning keywords at start of text
	planningKeywords := []struct {
		keyword string
		style   lipgloss.Style
	}{
		{"SCHEDULED:", r.styles.Scheduled},
		{"DEADLINE:", r.styles.Deadline},
		{"CLOSED:", r.styles.Closed},
	}

	for _, pk := range planningKeywords {
		if strings.HasPrefix(content, pk.keyword) {
			rest := content[len(pk.keyword):]
			// Check for inactive timestamp in the rest (for CLOSED)
			rest = r.renderInactiveTimestamps(rest)
			return pk.style.Render(pk.keyword) + rest
		}
		// Also check for keyword with leading space (e.g., " DEADLINE:")
		if strings.HasPrefix(content, " "+pk.keyword) {
			rest := content[len(pk.keyword)+1:]
			rest = r.renderInactiveTimestamps(rest)
			return " " + pk.style.Render(pk.keyword) + rest
		}
	}

	// Check for inactive timestamps anywhere in text
	return r.renderInactiveTimestamps(content)
}

// renderInactiveTimestamps finds and styles inactive timestamps [YYYY-MM-DD ...]
func (r *Renderer) renderInactiveTimestamps(content string) string {
	var result strings.Builder
	remaining := content

	for {
		// Find opening bracket for inactive timestamp
		start := strings.Index(remaining, "[")
		if start == -1 {
			result.WriteString(remaining)
			break
		}

		// Find closing bracket
		end := strings.Index(remaining[start:], "]")
		if end == -1 {
			result.WriteString(remaining)
			break
		}
		end += start // Adjust to absolute position

		// Check if this looks like an inactive timestamp [YYYY-MM-DD ...]
		timestampContent := remaining[start+1 : end]
		if len(timestampContent) >= 10 && isInactiveTimestamp(timestampContent) {
			result.WriteString(remaining[:start])
			result.WriteString(r.styles.Timestamp.Render("[" + timestampContent + "]"))
			remaining = remaining[end+1:]
		} else {
			// Not a timestamp, keep going
			result.WriteString(remaining[:end+1])
			remaining = remaining[end+1:]
		}
	}

	return result.String()
}

// isInactiveTimestamp checks if content looks like a timestamp (YYYY-MM-DD ...)
func isInactiveTimestamp(content string) bool {
	// Basic check: starts with date pattern YYYY-MM-DD
	if len(content) < 10 {
		return false
	}
	// Check for digit patterns at expected positions
	for i, c := range content[:10] {
		if i == 4 || i == 7 {
			if c != '-' {
				return false
			}
		} else {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func (r *Renderer) renderEmphasis(e goorg.Emphasis) string {
	content := r.renderInlineNodes(e.Content)

	// go-org uses the actual marker character as the Kind
	switch e.Kind {
	case "*":
		return r.styles.Bold.Render(content)
	case "/":
		return r.styles.Italic.Render(content)
	case "_":
		return r.styles.Underline.Render(content)
	case "=":
		return r.styles.Verbatim.Render(content)
	case "~":
		return r.styles.InlineCode.Render(content)
	case "+":
		return r.styles.Strikethrough.Render(content)
	default:
		return content
	}
}

func (r *Renderer) renderLink(link goorg.RegularLink) string {
	var text string
	if len(link.Description) > 0 {
		text = r.renderInlineNodes(link.Description)
	} else {
		text = link.URL
	}

	// Truncate long URLs for display
	displayText := text
	maxLen := 40
	if len(displayText) > maxLen {
		displayText = displayText[:maxLen-3] + "..."
	}

	// Determine link type and icon
	var icon string
	switch {
	case strings.HasPrefix(link.URL, "http://") || strings.HasPrefix(link.URL, "https://"):
		icon = "🔗"
	case strings.HasPrefix(link.URL, "file:"):
		icon = "📄"
	case strings.HasPrefix(link.URL, "mailto:"):
		icon = "📧"
	case strings.HasSuffix(link.URL, ".org"):
		icon = "📝"
	default:
		icon = "→"
	}

	return r.styles.Link.Render(icon + " " + displayText)
}

func (r *Renderer) renderTimestamp(ts goorg.Timestamp) string {
	// Format the timestamp nicely
	var formatted string
	if ts.IsDate {
		formatted = ts.Time.Format("2006-01-02 Mon")
	} else {
		formatted = ts.Time.Format("2006-01-02 Mon 15:04")
	}

	// Add repeater/interval if present
	if ts.Interval != "" {
		formatted += " " + ts.Interval
	}

	// Use calendar emoji and styled timestamp
	return r.styles.Timestamp.Render("📅 " + formatted)
}

func (r *Renderer) renderFootnoteLink(fn goorg.FootnoteLink) string {
	return r.styles.FootnoteRef.Render("[" + fn.Name + "]")
}

// extractBlockText extracts plain text from block children
func (r *Renderer) extractBlockText(nodes []goorg.Node) string {
	var b strings.Builder
	for _, node := range nodes {
		switch n := node.(type) {
		case goorg.Text:
			b.WriteString(n.Content)
		default:
			b.WriteString(fmt.Sprintf("%v", n))
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}
