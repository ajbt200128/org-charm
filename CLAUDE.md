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
| `github.com/charmbracelet/harmonica` | Spring-based animations |
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
- **Planning** (SCHEDULED, DEADLINE, CLOSED with distinct colors)
- **Paragraphs**
- **Lists** (unordered, ordered, definition lists, checklists, **nested lists**)
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
- **Active timestamps** (`<2024-01-01 Mon>`)
- **Inactive timestamps** (`[2024-01-01 Mon]`) - styled in text
- **Footnote references** (`[fn:1]`)
- **Statistics** (`[2/4]`, `[50%]`)

### Keybindings
- `r` - Toggle raw/rendered view in document view

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

## Important Implementation Details

### go-org AST Gotcha: List Items Wrap Content in Paragraphs

When parsing list items, go-org wraps the content in `Paragraph` nodes:

```org
- List item with *bold* text
```

Produces this AST:
```
ListItem
  └── Paragraph          ← block element!
        ├── Text("List item with ")
        ├── Emphasis("*bold*")
        └── Text(" text")
```

**Not this** (what you might expect):
```
ListItem
  ├── Text("List item with ")
  ├── Emphasis("*bold*")
  └── Text(" text")
```

So when rendering list items, you must extract inline nodes from `Paragraph.Children`:

```go
for _, child := range item.Children {
    switch c := child.(type) {
    case goorg.Paragraph:
        content += r.renderInlineNodes(c.Children)  // Extract from Paragraph
    case goorg.List:
        // Nested list - render recursively with increased indent
        nestedContent += r.renderListWithIndent(c, indent+1)
    default:
        content += r.RenderNode(child)
    }
}
```

### Planning Keywords Are Parsed as Text

Planning keywords (SCHEDULED:, DEADLINE:, CLOSED:) are NOT special nodes in go-org.
They appear as regular `Text` nodes within a `Paragraph` following a headline:

```
Headline
  └── Paragraph
        ├── Text("SCHEDULED: ")
        ├── Timestamp(<2024-01-15 Mon>)
        ├── Text(" DEADLINE: ")
        └── Timestamp(<2024-01-20 Sat>)
```

Note: Inactive timestamps `[...]` (used with CLOSED) are NOT parsed as Timestamp nodes -
they remain as plain text. Use `renderInactiveTimestamps()` to detect and style them.

### SSH Color Profile Must Be Forced

Lipgloss uses `termenv` for color detection, which often fails over SSH (no real TTY to query). You must explicitly force TrueColor:

```go
// In bubbletea middleware setup
bubbletea.MiddlewareWithColorProfile(teaHandler, termenv.TrueColor)

// When creating session renderer
renderer := bubbletea.MakeRenderer(sess)
renderer.SetColorProfile(termenv.TrueColor)
```

Without this, all styles render as plain text with no ANSI codes.

## Animations with Harmonica

The TUI uses `charmbracelet/harmonica` for smooth spring-based transition animations.

### Animation Types

1. **Wave Ripple** (`AnimWaveRipple`) - Plays on initial SSH connection
   - Blue radial wave expands from center of screen
   - Content is revealed as the wave passes through each position
   - Uses Tokyo Night blue palette for wave characters (`░▒▓`)

2. **Poof** (`AnimPoof`) - Plays when toggling raw/rendered view (`r` key)
   - Old content scatters into particles (`·∘°⋅✦✧∗⁕※`)
   - Particles reform into new content
   - Applied only to viewport; header/footer remain static

### Animation Architecture

```go
// Animation state in Model
animType       AnimationType  // AnimNone, AnimWaveRipple, or AnimPoof
animSpring     harmonica.Spring
animValue      float64        // Progress 0.0 to 1.0
animVelocity   float64        // Spring velocity
animTarget     float64        // Target value (1.0)
```

### Spring Parameters

```go
animFPS       = 60   // Frame rate
animFrequency = 2.5  // Lower = slower animation
animDamping   = 1.0  // Critically damped (no overshoot)
```

### ANSI Code Handling

Animations must handle styled content with ANSI escape sequences. The `stripANSI()` helper extracts visual content for position calculations, while the wave animation tracks escape sequences separately to preserve styling in revealed areas.

### Adding New Animations

1. Add animation type constant to `AnimationType`
2. Add state fields to `Model` if needed
3. Create `apply*Animation()` method in `model.go`
4. Trigger animation in `Update()` (set `animType`, reset `animValue`)
5. Apply animation in `View()` or within specific render functions
6. Animation auto-completes when `animValue > 0.95`

## Known Limitations

- No image rendering (requires sixel support)
- No search within documents
- No export functionality
- Limited table alignment support
