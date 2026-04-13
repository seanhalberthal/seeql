# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Column filter (`/`) — filter rows by case-insensitive substring match on the selected column
- Horizontal column scrolling — columns keep natural widths, h/l scrolls through them
- Column scroll indicator in footer ("cols N–M of T")
- Version display in status bar bottom-right corner
- CI workflow (format, tidy, vet, lint, test, build)
- Release workflow with auto-versioning from commit prefixes
- golangci-lint configuration

### Fixed
- Sidebar/results height mismatch caused by variable footer line count
- Streaming results column sizing — recalculate widths when first data page arrives
- Minimum column width raised to 10 chars for readability

### Changed
- DSN-first connection manager flow
- Two-pane layout with floating editor overlay
- Single adaptive theme using ANSI 16-colour palette
- Simplified config format (JSON, DSN-only connections)
- Dropped DuckDB adapter (pure Go only)
