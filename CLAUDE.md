# Org-Charm Development Guide

## Project Overview

Org-Charm is an SSH server that serves Emacs org-mode files with beautiful TUI rendering. It uses the Charm.sh ecosystem for a modern terminal experience.

## Architecture

```
org-charm/
├── main.go              # SSH server entry point (wish + bubbletea middleware)
├── org/
│   └── parser.go        # go-org wrapper for parsing .org files
├── ui/
│   ├── model.go         # Bubbletea TUI model (file browser + document viewer)
│   ├── render.go        # Org AST to styled string renderer
│   └── styles.go        # Lipgloss theme definitions (Tokyo Night palette)
└── orgfiles/            # Default org files directory
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/wish` | SSH server framework |
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/charmbracelet/log` | Structured logging |
| `github.com/niklasfasching/go-org/org` | Org-mode parser |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting for code blocks |

## Development Commands

```bash
# Enter development environment
devenv shell

# Build
go build -o org-charm .

# Run server
./org-charm -dir ./orgfiles -port 2222

# Connect (from another terminal)
ssh localhost -p 2222

# Run tests
go test ./...
```

## Org-Mode Syntax Support

### Block Elements
- **Headings** (h1-h4 with rainbow colors, TODO/DONE badges, priority, tags)
- **Paragraphs**
- **Lists** (unordered, ordered, definition lists, checklists)
- **Code blocks** (`#+BEGIN_SRC`) with chroma syntax highlighting
- **Quote blocks** (`#+BEGIN_QUOTE`)
- **Example blocks** (`#+BEGIN_EXAMPLE`)
- **Verse blocks** (`#+BEGIN_VERSE`)
- **Tables** with borders and header detection
- **Horizontal rules** (`-----`)
- **Drawers** and property drawers
- **Footnote definitions**

### Inline Elements
- **Bold** (`*text*`)
- **Italic** (`/text/`)
- **Underline** (`_text_`)
- **Strikethrough** (`+text+`)
- **Code** (`~text~`)
- **Verbatim** (`=text=`)
- **Links** (`[[url][description]]`)
- **Timestamps** (`<2024-01-01 Mon>`)
- **Footnote references** (`[fn:1]`)
- **Statistics** (`[2/4]`, `[50%]`)

## go-org AST Types

The go-org library parses org files into an AST. Key types:

```go
// Block-level
goorg.Headline     // Has: Lvl, Status, Priority, Title, Tags, Children
goorg.Paragraph    // Has: Children (inline nodes)
goorg.Block        // Has: Name, Parameters, Children
goorg.List         // Has: Kind, Items
goorg.ListItem     // Has: Bullet, Status, Children
goorg.Table        // Has: Rows (each with Columns)

// Inline
goorg.Text         // Has: Content string
goorg.Emphasis     // Has: Kind ("*", "/", "_", "=", "~", "+"), Content
goorg.RegularLink  // Has: URL, Description
goorg.Timestamp    // Has: Time, IsDate, Interval
goorg.FootnoteLink // Has: Name
```

## Bubbletea Architecture

The TUI uses the Elm architecture:
- `Model` holds all state (view mode, selected file, viewport position)
- `Update` handles messages (key presses, window resizes)
- `View` renders the current state to a string

Key message types:
- `tea.KeyMsg` - keyboard input
- `tea.WindowSizeMsg` - terminal resize events

## Wish SSH Server

The server uses wish middleware pattern:
```go
wish.WithMiddleware(
    bubbletea.Middleware(teaHandler),  // TUI per session
    activeterm.Middleware(),            // Require PTY
    logging.Middleware(),               // Log connections
)
```

Each SSH connection gets its own bubbletea program instance with session-specific renderer.

## Styling Guidelines

Colors follow the Tokyo Night palette:
- `colorH1` (#f7768e) - Red for h1
- `colorH2` (#ff9e64) - Orange for h2
- `colorH3` (#e0af68) - Yellow for h3
- `colorH4` (#9ece6a) - Green for h4
- `colorHighlight` (#7aa2f7) - Blue for links, selected items
- `colorAccent` (#bb9af7) - Magenta for tags, quotes

## Adding New Org Elements

1. Add the node type case in `render.go` `RenderNode()` or `renderInlineNode()`
2. Create a `render*` method for the element
3. Add any new styles to `styles.go` (both struct field and initialization)
4. Test with sample org files

## Known Limitations

- No image rendering (requires sixel support)
- No search within documents
- No export functionality
- Limited table alignment support
