# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Cell value popover: press `P` on a focused result cell to open a scrollable, searchable dialog showing the full value. Auto-pretty-prints JSON, supports `/` search with `n`/`N`, and `y`/`Y` to yank displayed/raw value.
- Sidebar: `F5` / `Ctrl+G` / `Ctrl+Enter` executes `SELECT * FROM <table>` for the highlighted table node in the current tab.
- Column type is now shown next to the column name in the results-pane footer preview and in the cell popover title.
- Help overlay: two-column layout so all sections fit on shorter terminals.
- History browser: vim-style navigation (`j`/`k`, `g`/`G`, `Ctrl+d`/`Ctrl+u`)
- History browser: `e` loads the selected query into the editor and copies it to the clipboard
- History browser: `y` yanks the selected query to the clipboard
- History browser: filter is now gated behind `/` so nav keys stay free
- History browser: restyled to match the floating editor (focused border, wider layout)
- Column filter (`/`) — filter rows by case-insensitive substring match on the selected column
- Horizontal column scrolling — columns keep natural widths, h/l scrolls through them
- Column scroll indicator in footer ("cols N–M of T")
- Version display in status bar bottom-right corner
- CI workflow (format, tidy, vet, lint, test, build)
- Release workflow with auto-versioning from commit prefixes
- golangci-lint configuration

### Fixed
- Runtime panic when executing a second query whose result has more columns than the previous one (results table now clears rows before swapping in new column definitions).
- Cell popover: word-aware line wrapping so values no longer break mid-word; long unbroken tokens still hard-break to fit the dialog.
- History browser: long entries with wide metadata no longer wrap onto a second line
- Sidebar/results height mismatch caused by variable footer line count
- Streaming results column sizing — recalculate widths when first data page arrives
- Minimum column width raised to 10 chars for readability
- Autocomplete no longer pops up immediately after typing `'` or `"` (e.g. inside `WHERE name = 'sean'`); also restores `?` toggling the help overlay outside the editor

### Changed
- DSN-first connection manager flow
- Two-pane layout with floating editor overlay
- Single adaptive theme using ANSI 16-colour palette
- Simplified config format (JSON, DSN-only connections)
- Dropped DuckDB adapter (pure Go only)
