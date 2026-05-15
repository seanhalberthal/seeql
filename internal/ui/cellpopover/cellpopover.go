// Package cellpopover renders the focused result cell's full value in a
// scrollable, searchable, centred dialog. Activated with "P" in the results
// pane; designed for inspecting large jsonb / text values that are truncated
// in the inline cell preview.
package cellpopover

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/seanhalberthal/seeql/internal/theme"
)

// Model is the cell popover.
type Model struct {
	visible   bool
	width     int
	height    int
	colName   string
	colType   string   // column type from adapter (e.g. "text", "jsonb", "int4")
	rawValue  string   // original cell value
	displayed string   // value as shown (pretty-printed when JSON)
	pretty    bool     // whether displayed != rawValue
	lines     []string // soft-wrapped display lines
	wrapWidth int      // width used for the current wrap

	// View state
	offset int // index of the first visible display line

	// Search state
	search     textinput.Model
	searchMode bool
	matches    []match // line index + byte range of each match (in display lines)
	matchIdx   int     // index into matches for the currently-highlighted hit
}

// match locates a single search hit inside a wrapped display line.
type match struct {
	line  int // index into m.lines
	start int // byte offset in the line
	end   int // byte offset in the line
}

// New constructs a Model with default styling.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/ "
	ti.Width = 40
	return Model{search: ti}
}

// Visible reports whether the popover is currently shown.
func (m Model) Visible() bool { return m.visible }

// SetSize stores the terminal dimensions used for centring and wrap width.
func (m *Model) SetSize(w, h int) {
	if m.width == w && m.height == h {
		return
	}
	m.width = w
	m.height = h
	if m.visible {
		m.rewrap()
	}
}

// Show opens the popover for the given cell value. Auto-pretty-prints JSON.
func (m *Model) Show(columnName, columnType, value string) {
	m.visible = true
	m.colName = columnName
	m.colType = columnType
	m.rawValue = value
	m.displayed, m.pretty = prettyJSON(value)
	m.offset = 0
	m.matchIdx = 0
	m.matches = nil
	m.searchMode = false
	m.search.SetValue("")
	m.search.Blur()
	m.rewrap()
}

// Hide closes the popover and clears search state.
func (m *Model) Hide() {
	m.visible = false
	m.searchMode = false
	m.search.Blur()
}

// Update handles all key input while the popover is visible.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.searchMode {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	if m.searchMode {
		return m.updateSearchMode(keyMsg)
	}
	return m.updateNavMode(keyMsg)
}

// updateNavMode handles keys while the content has focus (no search input).
func (m Model) updateNavMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	visH := m.visibleLineCount()
	maxOffset := m.maxOffset()

	switch msg.String() {
	case "esc", "q":
		m.visible = false
		return m, nil

	case "/":
		m.searchMode = true
		m.search.Focus()
		return m, textinput.Blink

	case "y":
		_ = clipboard.WriteAll(m.displayed)
		return m, nil

	case "Y":
		_ = clipboard.WriteAll(m.rawValue)
		return m, nil

	case "up", "k", "ctrl+p":
		if m.offset > 0 {
			m.offset--
		}
		return m, nil

	case "down", "j", "ctrl+n":
		if m.offset < maxOffset {
			m.offset++
		}
		return m, nil

	case "g", "home":
		m.offset = 0
		return m, nil

	case "G", "end":
		m.offset = maxOffset
		return m, nil

	case "pgup", "ctrl+u":
		m.offset -= visH
		if m.offset < 0 {
			m.offset = 0
		}
		return m, nil

	case "pgdown", "ctrl+d", " ":
		m.offset += visH
		if m.offset > maxOffset {
			m.offset = maxOffset
		}
		return m, nil

	case "n":
		if len(m.matches) > 0 {
			m.matchIdx = (m.matchIdx + 1) % len(m.matches)
			m.ensureMatchVisible()
		}
		return m, nil

	case "N":
		if len(m.matches) > 0 {
			m.matchIdx = (m.matchIdx - 1 + len(m.matches)) % len(m.matches)
			m.ensureMatchVisible()
		}
		return m, nil
	}
	return m, nil
}

// updateSearchMode handles keys while the search input is focused.
func (m Model) updateSearchMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.search.Blur()
		m.search.SetValue("")
		m.matches = nil
		return m, nil

	case "enter":
		m.searchMode = false
		m.search.Blur()
		m.findMatches()
		if len(m.matches) > 0 {
			m.matchIdx = 0
			m.ensureMatchVisible()
		}
		return m, nil
	}

	prev := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != prev {
		m.findMatches()
		if len(m.matches) > 0 {
			m.matchIdx = 0
			m.ensureMatchVisible()
		}
	}
	return m, cmd
}

// View renders the popover. Returns "" when not visible.
func (m Model) View() string {
	if !m.visible {
		return ""
	}
	th := theme.Current
	w, _ := m.dialogSize()
	innerW := w - 2

	titleText := fmt.Sprintf("Cell: %s", m.colName)
	title := th.DialogTitle.Render(titleText)
	if m.colType != "" {
		title += "  " + th.MutedText.Render(m.colType)
	}

	var searchLine string
	switch {
	case m.searchMode:
		searchLine = m.search.View()
	case m.search.Value() != "" && len(m.matches) > 0:
		searchLine = th.MutedText.Render(fmt.Sprintf("search: %s  [%d/%d]",
			m.search.Value(), m.matchIdx+1, len(m.matches)))
	case m.search.Value() != "" && len(m.matches) == 0:
		searchLine = th.ErrorText.Render(fmt.Sprintf("search: %s  (no matches)", m.search.Value()))
	default:
		searchLine = th.MutedText.Render("press / to search")
	}

	visH := m.visibleLineCount()
	end := m.offset + visH
	if end > len(m.lines) {
		end = len(m.lines)
	}
	var body []string
	for i := m.offset; i < end; i++ {
		body = append(body, m.renderLine(i, th))
	}
	for len(body) < visH {
		body = append(body, "")
	}

	counter := fmt.Sprintf("line %d-%d of %d", m.offset+1, end, len(m.lines))
	if len(m.lines) == 0 {
		counter = "(empty)"
	}
	if m.pretty {
		counter += "  •  JSON pretty-printed"
	}

	var hint string
	if m.searchMode {
		hint = th.MutedText.Render("enter: jump  esc: cancel")
	} else {
		hint = th.MutedText.Render("j/k: scroll  g/G: top/bottom  /: search  n/N: next/prev  y: yank  esc: close")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		searchLine,
		"",
		strings.Join(body, "\n"),
		"",
		th.MutedText.Render(counter),
		hint,
	)

	return th.FocusedBorder.
		Width(innerW).
		Padding(0, 1).
		Render(content)
}

// renderLine renders one wrapped display line, highlighting any active search
// matches. The currently-selected match (m.matchIdx) is rendered with extra
// emphasis.
func (m Model) renderLine(i int, th *theme.Theme) string {
	line := m.lines[i]
	if len(m.matches) == 0 {
		return line
	}
	var b strings.Builder
	cursor := 0
	for mi, mt := range m.matches {
		if mt.line != i {
			continue
		}
		if mt.start > cursor {
			b.WriteString(line[cursor:mt.start])
		}
		segment := line[mt.start:mt.end]
		if mi == m.matchIdx {
			b.WriteString(th.SidebarSelected.Render(segment))
		} else {
			b.WriteString(th.SuccessText.Render(segment))
		}
		cursor = mt.end
	}
	if cursor < len(line) {
		b.WriteString(line[cursor:])
	}
	return b.String()
}

// dialogSize returns the outer width/height for the popover dialog.
func (m Model) dialogSize() (int, int) {
	w := m.width * 3 / 4
	if w < 60 {
		w = m.width - 4
	}
	if w > m.width-4 {
		w = m.width - 4
	}
	if w < 20 {
		w = 20
	}

	h := m.height * 3 / 5
	if h < 15 {
		h = m.height - 4
	}
	if h > m.height-4 {
		h = m.height - 4
	}
	if h < 10 {
		h = 10
	}
	return w, h
}

// visibleLineCount returns how many wrapped display lines fit in the body
// area, accounting for chrome (border + title + search + blanks + counter
// + hint).
func (m Model) visibleLineCount() int {
	_, h := m.dialogSize()
	avail := h - 8
	if avail < 3 {
		avail = 3
	}
	return avail
}

// maxOffset returns the largest valid offset such that visible window is full.
func (m Model) maxOffset() int {
	max := len(m.lines) - m.visibleLineCount()
	if max < 0 {
		max = 0
	}
	return max
}

// rewrap recalculates the soft-wrapped display lines for the current dialog
// width and re-runs the search to fix up match line indices.
func (m *Model) rewrap() {
	w, _ := m.dialogSize()
	wrapW := w - 4
	if wrapW < 10 {
		wrapW = 10
	}
	m.wrapWidth = wrapW
	m.lines = wrapLines(m.displayed, wrapW)
	if m.offset > m.maxOffset() {
		m.offset = m.maxOffset()
	}
	if m.search.Value() != "" {
		m.findMatches()
		if m.matchIdx >= len(m.matches) {
			m.matchIdx = 0
		}
	}
}

// findMatches recomputes m.matches for the current search term against
// m.lines. Case-insensitive substring search.
func (m *Model) findMatches() {
	m.matches = nil
	needle := strings.ToLower(m.search.Value())
	if needle == "" {
		return
	}
	for li, line := range m.lines {
		lower := strings.ToLower(line)
		from := 0
		for {
			idx := strings.Index(lower[from:], needle)
			if idx < 0 {
				break
			}
			start := from + idx
			end := start + len(needle)
			m.matches = append(m.matches, match{line: li, start: start, end: end})
			from = end
			if end >= len(line) {
				break
			}
		}
	}
}

// ensureMatchVisible scrolls the viewport so the current match line is
// inside the visible window.
func (m *Model) ensureMatchVisible() {
	if m.matchIdx < 0 || m.matchIdx >= len(m.matches) {
		return
	}
	target := m.matches[m.matchIdx].line
	visH := m.visibleLineCount()
	switch {
	case target < m.offset:
		m.offset = target
	case target >= m.offset+visH:
		m.offset = target - visH + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > m.maxOffset() {
		m.offset = m.maxOffset()
	}
}

// prettyJSON returns (pretty, true) when s parses as JSON; otherwise (s, false).
// Only triggers when the trimmed value starts with { or [ to avoid wasting
// cycles on every cell.
func prettyJSON(s string) (string, bool) {
	t := strings.TrimSpace(s)
	if len(t) == 0 || (t[0] != '{' && t[0] != '[') {
		return s, false
	}
	var v any
	if err := json.Unmarshal([]byte(t), &v); err != nil {
		return s, false
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s, false
	}
	return string(out), true
}

// wrapLines splits s into display lines, soft-wrapping any source line wider
// than width. Prefers breaking at whitespace; falls back to a hard
// character-level break for tokens that don't fit on a single line (e.g. long
// UUIDs). Uses display-width (runewidth) so multi-byte characters wrap
// correctly. Tab characters are expanded to two spaces.
func wrapLines(s string, width int) []string {
	if width <= 0 {
		width = 1
	}
	src := strings.Split(strings.ReplaceAll(s, "\t", "  "), "\n")
	var out []string
	for _, line := range src {
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapOneLine(line, width)...)
	}
	return out
}

// wrapOneLine word-wraps a single source line. Splits on whitespace runs,
// preserving them, and starts a new line when the next chunk would overflow.
// Tokens longer than width are hard-broken at character boundaries.
func wrapOneLine(line string, width int) []string {
	tokens := tokenize(line)
	var (
		out []string
		cur strings.Builder
		w   int
	)
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
			w = 0
		}
	}
	for _, tok := range tokens {
		tw := runewidth.StringWidth(tok)
		if isSpace(tok) {
			if w == 0 {
				continue
			}
			if w+tw > width {
				flush()
				continue
			}
			cur.WriteString(tok)
			w += tw
			continue
		}
		if w+tw <= width {
			cur.WriteString(tok)
			w += tw
			continue
		}
		flush()
		if tw <= width {
			cur.WriteString(tok)
			w = tw
			continue
		}
		for _, r := range tok {
			rw := runewidth.RuneWidth(r)
			if w+rw > width {
				flush()
			}
			cur.WriteRune(r)
			w += rw
		}
	}
	flush()
	if len(out) == 0 {
		out = append(out, "")
	}
	return out
}

// tokenize splits s into alternating whitespace and non-whitespace runs.
func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	curSpace := false
	for i, r := range s {
		sp := r == ' '
		if i == 0 {
			curSpace = sp
			cur.WriteRune(r)
			continue
		}
		if sp == curSpace {
			cur.WriteRune(r)
			continue
		}
		out = append(out, cur.String())
		cur.Reset()
		cur.WriteRune(r)
		curSpace = sp
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func isSpace(s string) bool { return len(s) > 0 && s[0] == ' ' }
