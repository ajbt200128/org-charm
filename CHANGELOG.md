# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nested footnote support with distinct symbols and indentation per level
  - Level 0: [1], [2], [3] (yellow)
  - Level 1: a., b., c. (cyan)
  - Level 2: i., ii., iii. (magenta)
  - Level 3: α, β, γ (orange)

## [0.2.0] - 2026-02-26

### Added
- Spring-based transition animations using `charmbracelet/harmonica`
  - Wave ripple animation on initial SSH connection (blue radial wave reveals content)
  - Poof animation when toggling raw/rendered view (particles scatter and reform)
- Index.org support for custom main page headers
- Document cycling with `n`/`p` keys to navigate between files
- Author metadata display in file list and document headers
- Raw view toggle (`r` key) to see original org-mode source
- Nested list support with proper indentation
- Planning timestamps (SCHEDULED, DEADLINE, CLOSED) with distinct colors
- Syntax highlighting for code blocks via Chroma
- Tokyo Night color palette throughout the UI
- Docker support with multi-arch images (linux/amd64, linux/arm64)
- GitHub Actions CI/CD pipeline

### Fixed
- Inline formatting in list items (bold, italic, code, etc.)
- SSH color profile detection (force TrueColor for proper styling)
- ANSI escape sequence handling in animations

### Technical
- Bubbletea TUI framework with Elm architecture
- Wish SSH server with per-session renderers
- go-org parser for org-mode AST
- Lipgloss for terminal styling

[0.2.0]: https://github.com/ajbt200128/org-charm/releases/tag/v0.2.0
