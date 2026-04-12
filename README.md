<div align="center">

<picture>
  <img alt="seeql" src=".github/assets/logo.svg" width="280">
</picture>

<br>

**Terminal SQL client for PostgreSQL, MySQL, and SQLite.**

[![Go 1.25+](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Licence](https://img.shields.io/github/license/seanhalberthal/seeql?style=flat&label=licence&color=4493F8)](LICENCE)
[![macOS](https://img.shields.io/badge/macOS-supported-6e7681?style=flat&logo=apple&logoColor=white)]()
[![Linux](https://img.shields.io/badge/Linux-supported-6e7681?style=flat&logo=linux&logoColor=white)]()

[Quick Start](#quick-start) · [Features](#features) · [Keybindings](#keybindings) · [Configuration](#configuration)

</div>

---

## Quick Start

```bash
brew install seanhalberthal/tap/seeql
```

```bash
seeql                                    # Launch connection manager
seeql postgres://user:pass@host/db       # Connect via DSN
seeql ./data.db                          # SQLite file
```

---

## Features

- **Two-pane layout** — sidebar + results table. Query editor floats on demand.
- **Terminal-native theme** — inherits your terminal's colour scheme. No hardcoded colours.
- **Streaming results** — constant memory for arbitrarily large result sets.
- **Context-aware autocomplete** — tables, columns, keywords, functions.
- **Multi-tab** — each tab owns its result set and query.
- **Connection manager** — DSN-based, adapter auto-detected.
- **Query history** — SQLite-backed, searchable via Ctrl+H.
- **Audit log** — opt-in JSON Lines trail for compliance.
- **Pure Go** — zero CGo, cross-platform, instant startup.

---

## Install

### Homebrew

```bash
brew install seanhalberthal/tap/seeql
```

### From source

```bash
go install github.com/seanhalberthal/seeql/cmd/seeql@latest
```

### Build from repo

```bash
git clone https://github.com/seanhalberthal/seeql.git
cd seeql
make build
# Binary at bin/seeql
```

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus: sidebar / results |
| `e` | Open query editor |
| `Ctrl+S` | Toggle sidebar |
| `Ctrl+O` | Connection manager |
| `Ctrl+R` | Refresh schema |
| `Ctrl+E` | Export results to CSV |
| `?` / `F1` | Help |
| `q` / `Ctrl+Q` | Quit |

### Editor (floating)

| Key | Action |
|-----|--------|
| `F5` / `Ctrl+G` | Execute query |
| `Ctrl+C` | Cancel running query |
| `Ctrl+H` | Query history |
| `Esc` | Close editor |

### Tabs

| Key | Action |
|-----|--------|
| `Ctrl+T` | New tab |
| `X` | Close tab |
| `Ctrl+]` / `Ctrl+[` | Next / previous tab |

---

## Configuration

Config at `~/.config/seeql/config.json`:

```json
{
  "keymode": "vim",
  "editor": {
    "tab_size": 4,
    "show_line_numbers": true
  },
  "results": {
    "page_size": 1000,
    "max_column_width": 50
  },
  "connections": [
    {
      "name": "local pg",
      "dsn": "postgres://user:pass@localhost:5432/mydb"
    },
    {
      "dsn": "./data.db"
    }
  ]
}
```

---

## Supported Databases

| Database | Driver | CGo |
|----------|--------|-----|
| PostgreSQL | [pgx](https://github.com/jackc/pgx) | No |
| MySQL | [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) | No |
| SQLite | [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | No |

---

<details>
<summary><strong>Architecture</strong></summary>

```
seeql/
├── cmd/seeql/              # CLI entry point
├── internal/
│   ├── adapter/            # Database adapter interface + drivers
│   │   ├── postgres/
│   │   ├── mysql/
│   │   └── sqlite/
│   ├── app/                # Root Bubble Tea model
│   ├── ui/                 # UI components
│   │   ├── sidebar/        # Schema tree browser
│   │   ├── editor/         # SQL editor + syntax highlighting
│   │   ├── results/        # Results table + exporter
│   │   ├── tabs/           # Tab bar
│   │   ├── statusbar/      # Status bar
│   │   ├── autocomplete/   # Autocomplete dropdown
│   │   ├── connmgr/        # Connection manager
│   │   ├── historybrowser/  # Query history overlay
│   │   └── dialog/         # Reusable dialog
│   ├── completion/         # SQL completion engine
│   ├── schema/             # Schema types
│   ├── config/             # JSON config
│   ├── history/            # Query history (SQLite)
│   ├── audit/              # JSON Lines audit log
│   └── theme/              # Adaptive theme (ANSI 16-colour)
├── Makefile
└── .goreleaser.yaml
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

</details>

---

## Licence

[MIT](LICENCE)
