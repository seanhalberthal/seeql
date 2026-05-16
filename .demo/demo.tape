# seeql demo — produces .demo/demo.gif and .demo/demo.mp4 for the README.
#
# Prereqs:
#   make build                              # produces ./seeql with version info
#   sqlite3 .demo/demo.db < .demo/demo.sql  # if .demo/demo.db isn't there yet
#
# Run from the repo root:
#   vhs .demo/demo.tape

Output .demo/demo.gif
Output .demo/demo.mp4

Set Shell "zsh"
Set FontFamily "JetBrainsMono Nerd Font Mono"
Set FontSize 13
Set Width 1280
Set Height 760
Set Theme { "name": "Monokai", "background": "#272822", "foreground": "#f8f8f2", "cursor": "#f8f8f0", "selection": "#49483e", "black": "#272822", "red": "#f92672", "green": "#a6e22e", "yellow": "#f4bf75", "blue": "#66d9ef", "magenta": "#ae81ff", "cyan": "#a1efe4", "white": "#f8f8f2", "brightBlack": "#75715e", "brightRed": "#f92672", "brightGreen": "#a6e22e", "brightYellow": "#f4bf75", "brightBlue": "#66d9ef", "brightMagenta": "#ae81ff", "brightCyan": "#a1efe4", "brightWhite": "#f9f8f5" }
Set TypingSpeed 70ms
Set PlaybackSpeed 1.0

# ── Clean prompt (hidden) ────────────────────────────────────
Hide
Type "cd seeql"
Enter
Type "clear"
Enter
Type "export SEEQL_HIDE_VERSION=1"
Enter
Show

# ── Launch ───────────────────────────────────────────────────
Type "./seeql .demo/demo.db"
Enter
Sleep 2.5s

# ── First query: products with JSON metadata ─────────────────
Type "e"
Sleep 600ms

Type "SELECT name, category, price, metadata"
Enter
Type "FROM products"
Enter
Type "ORDER BY price DESC"
Enter
Type "LIMIT 20;"
Sleep 1.3s

Ctrl+G
Sleep 2.5s

# ── Sidebar toggle + expand (Ctrl+S / Ctrl+F) ────────────────
# Hide sidebar — results table claims the full width
Ctrl+S
Sleep 1.6s

# Show sidebar again
Ctrl+S
Sleep 1.2s

# Expand sidebar to half-width to reveal the full schema tree
Ctrl+F
Sleep 2s

# Restore sidebar to its normal width
Ctrl+F
Sleep 1s

# ── Vim column navigation (h/l) ──────────────────────────────
Type "l"
Sleep 450ms
Type "l"
Sleep 450ms
Type "l"
Sleep 600ms
Type "h"
Sleep 500ms
Type "l"
Sleep 800ms

# ── Yank cell to clipboard (`y` → "Yanked <col> to clipboard") ─
Type "y"
Sleep 1.6s

# ── Cell popover on the JSON metadata column (P) ─────────────
Type "P"
Sleep 2.5s

# Yank pretty-printed value from inside the popover
Type "y"
Sleep 1.5s

Escape
Sleep 700ms

# ── New tab + aggregate query ────────────────────────────────
Ctrl+T
Sleep 700ms
Type "e"
Sleep 500ms

Type "SELECT status,"
Enter
Type "       COUNT(*) AS count,"
Enter
Type "       ROUND(SUM(total), 2) AS revenue"
Enter
Type "FROM orders"
Enter
Type "GROUP BY status"
Enter
Type "ORDER BY revenue DESC;"
Sleep 1.3s

Ctrl+G
Sleep 2.5s

# ── History browser (Ctrl+H) ─────────────────────────────────
Ctrl+H
Sleep 1.5s

# Vim navigation through history
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "k"
Sleep 700ms

# Yank highlighted history entry to clipboard
Type "y"
Sleep 1.6s

Escape
Sleep 800ms

# ── Quit ─────────────────────────────────────────────────────
Ctrl+Q
Sleep 600ms
