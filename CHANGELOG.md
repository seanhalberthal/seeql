# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.0.4] - 2026-05-15

### Added
- Cell value popover: press `P` on a focused result cell to open a scrollable, searchable dialog showing the full value. Auto-pretty-prints JSON, supports `/` search with `n`/`N`, and `y`/`Y` to yank displayed/raw value.
- Sidebar: `F5` / `Ctrl+G` / `Ctrl+Enter` executes `SELECT * FROM <table>` for the highlighted table node in the current tab.
- Column type is now shown next to the column name in the results-pane footer preview and in the cell popover title.
- Help overlay: two-column layout so all sections fit on shorter terminals.

### Fixed
- Runtime panic when executing a second query whose result has more columns than the previous one (results table now clears rows before swapping in new column definitions).
- Cell popover: word-aware line wrapping so values no longer break mid-word; long unbroken tokens still hard-break to fit the dialog.

## [0.0.3] - 2026-05-15

### Fixed
- Autocomplete no longer pops up immediately after typing `'` or `"` (e.g. inside `WHERE name = 'sean'`); also restores `?` toggling the help overlay outside the editor.

## [0.0.2] - 2026-04-14

### Added
- History browser: vim-style navigation (`j`/`k`, `g`/`G`, `Ctrl+d`/`Ctrl+u`).
- History browser: `e` loads the selected query into the editor and copies it to the clipboard.
- History browser: `y` yanks the selected query to the clipboard.
- History browser: filter is now gated behind `/` so nav keys stay free.
- History browser: restyled to match the floating editor (focused border, wider layout).

### Fixed
- History browser: long entries with wide metadata no longer wrap onto a second line.

## [0.0.1] - 2026-04-14

Initial release of seeql (forked from `sadopc/gotermsql`).

### Added
- Two-pane TUI layout (sidebar + results) with a floating editor overlay.
- DSN-first connection manager with auto-detection of adapter from DSN.
- Single adaptive theme using the ANSI 16-colour palette so seeql inherits the terminal's colour scheme.
- Pure-Go adapters for PostgreSQL, MySQL, and SQLite, with batch schema introspection where available.
- Streaming SELECT results with pagination and a 5000-row sliding-window buffer (constant memory on 10M-row scans).
- Multi-tab editor/results state with per-tab `RunID` and connection-generation guards against stale async messages.
- Autocomplete engine with fuzzy matching, context-aware suggestions (FROM/SELECT/dot), and qualified `table.column` lookup.
- Syntax-highlighted SQL editor (Chroma) with autocomplete dropdown.
- Query history browser (`Ctrl+H`) backed by SQLite at `~/.config/seeql/history.db`.
- Full-screen help overlay with all keybindings (`?` / `F1`).
- Opt-in JSON Lines audit log for compliance, with DSN credential sanitisation and size-based rotation.
- CSV/JSON export, including streaming export for large result sets; `Ctrl+E` exports current results to CSV.
- Column filter (`/`) — filter rows by case-insensitive substring match on the selected column.
- Horizontal column scrolling — columns keep natural widths, `h`/`l` scrolls through them.
- Column scroll indicator in footer ("cols N–M of T").
- Zebra-striped results table and `Ctrl+J` focus binding.
- Version display in status bar bottom-right corner.
- Neovim integration support: `SEEQL_HEIGHT_OFFSET` env var and non-emoji sidebar icons when `NVIM` is set, to work around libvterm quirks.
- PostgreSQL integration test suite (auto-skips when no instance is available).
- CI workflow (format, tidy, vet, lint, test, build).
- Release workflow with auto-versioning from commit prefixes, cross-compiled binaries (linux/darwin/windows), and Homebrew tap.
- `golangci-lint` configuration.

### Changed
- DSN-first connection manager flow (replaced per-field host/port/user/password inputs).
- Simplified config format: JSON at `~/.config/seeql/config.json`, DSN-only saved connections.
- Pane-switch hint shown as `Shift+Tab` in the status bar and help overlay.
- Sidebar scrolling: highlight now moves independently of the viewport.

### Fixed
- Sidebar/results height mismatch caused by variable footer line count.
- Streaming results column sizing — recalculate widths when the first data page arrives.
- Minimum column width raised to 10 chars for readability.
- Results table columns no longer run together — cell style includes `Padding(0, 1)` and width calc accounts for the overhead.
- Status bar auto-clears 5 seconds after query results/errors so key hints stay visible.
- Tab-switch focus, reconnect resource leak, connmgr `File` field, and status bar timer reset bugs.
- Multiple critical nil panics, resource leaks, and staleness bugs (stale results across tabs, MySQL cancel, schema N+1, SQL identifier quoting, query timeouts, row caps, memory bounds).
- Numerous TUI rendering bugs surfaced during integration testing.

### Removed
- DuckDB adapter dropped — pure Go only (PostgreSQL, MySQL, SQLite).
