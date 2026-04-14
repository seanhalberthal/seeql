package historybrowser

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seanhalberthal/seeql/internal/history"
	"github.com/seanhalberthal/seeql/internal/theme"
)

// SelectQueryMsg is sent when the user picks a history entry to run immediately.
type SelectQueryMsg struct {
	Query string
}

// LoadQueryMsg is sent when the user copies a history entry to the clipboard
// and loads it into the editor without executing.
type LoadQueryMsg struct {
	Query string
}

// ClearYankMsg clears the transient "yanked to clipboard" hint.
type ClearYankMsg struct {
	Gen uint64
}

// Model is the history browser modal.
type Model struct {
	hist       *history.History
	entries    []history.HistoryEntry
	cursor     int
	offset     int // scroll offset
	visible    bool
	width      int
	height     int
	search     textinput.Model
	searchMode bool
	yankMsg    string
	yankGen    uint64
}

// New creates a new history browser.
func New(hist *history.History) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter queries..."
	ti.Prompt = "/ "
	ti.Width = 50
	return Model{
		hist:   hist,
		search: ti,
	}
}

// Show makes the history browser visible and loads entries.
func (m *Model) Show() {
	m.visible = true
	m.cursor = 0
	m.offset = 0
	m.searchMode = false
	m.search.SetValue("")
	m.search.Blur()
	m.loadEntries()
}

// Hide hides the history browser.
func (m *Model) Hide() {
	m.visible = false
	m.searchMode = false
	m.search.Blur()
}

// Visible returns whether the history browser is shown.
func (m Model) Visible() bool { return m.visible }

// SetSize sets the available space.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles history browser messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case ClearYankMsg:
		if msg.Gen == m.yankGen {
			m.yankMsg = ""
		}
		return m, nil
	case tea.KeyMsg:
		if m.searchMode {
			return m.updateSearchMode(msg)
		}
		return m.updateNavMode(msg)
	}

	// Non-key messages (e.g. blink) only matter when the input has focus.
	if m.searchMode {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}
	return m, nil
}

// copyToClipboard writes the query to the clipboard, records a transient hint,
// and returns a command that clears the hint after 2s.
func (m *Model) copyToClipboard(query string) tea.Cmd {
	_ = clipboard.WriteAll(query)
	m.yankGen++
	m.yankMsg = "Yanked to clipboard"
	gen := m.yankGen
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return ClearYankMsg{Gen: gen}
	})
}

// updateNavMode handles keys while the list has focus (no search input).
func (m Model) updateNavMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+h", "q":
		m.visible = false
		return m, nil

	case "/":
		m.searchMode = true
		m.search.Focus()
		return m, textinput.Blink

	case "up", "k", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, nil

	case "down", "j", "ctrl+n":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
			m.ensureVisible()
		}
		return m, nil

	case "g", "home":
		m.cursor = 0
		m.ensureVisible()
		return m, nil

	case "G", "end":
		m.cursor = len(m.entries) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case "pgup", "ctrl+u":
		visible := m.visibleCount()
		m.cursor -= visible
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case "pgdown", "ctrl+d":
		visible := m.visibleCount()
		m.cursor += visible
		if m.cursor >= len(m.entries) {
			m.cursor = len(m.entries) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureVisible()
		return m, nil

	case "enter":
		if m.cursor < len(m.entries) {
			query := m.entries[m.cursor].Query
			m.visible = false
			return m, func() tea.Msg {
				return SelectQueryMsg{Query: query}
			}
		}
		return m, nil

	case "e":
		if m.cursor < len(m.entries) {
			query := m.entries[m.cursor].Query
			_ = clipboard.WriteAll(query)
			m.visible = false
			return m, func() tea.Msg {
				return LoadQueryMsg{Query: query}
			}
		}
		return m, nil

	case "y":
		if m.cursor < len(m.entries) {
			return m, m.copyToClipboard(m.entries[m.cursor].Query)
		}
		return m, nil
	}
	return m, nil
}

// updateSearchMode handles keys while the filter input is focused.
func (m Model) updateSearchMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Leave search mode and clear the filter.
		m.searchMode = false
		m.search.Blur()
		if m.search.Value() != "" {
			m.search.SetValue("")
			m.cursor = 0
			m.offset = 0
			m.loadEntries()
		}
		return m, nil

	case "enter", "down", "up":
		// Commit the filter and return to nav mode so j/k/e/enter work.
		m.searchMode = false
		m.search.Blur()
		return m, nil
	}

	prevVal := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != prevVal {
		m.cursor = 0
		m.offset = 0
		m.loadEntries()
	}
	return m, cmd
}

// View renders the history browser styled like the floating editor.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	th := theme.Current
	w, _ := m.dialogSize()
	innerW := w - 2 // account for Padding(0, 1)

	title := th.DialogTitle.Render("Query History")

	var searchLine string
	if m.searchMode {
		searchLine = m.search.View()
	} else if v := m.search.Value(); v != "" {
		searchLine = th.MutedText.Render("filter: "+v) + "  " + th.MutedText.Render("(press / to edit)")
	} else {
		searchLine = th.MutedText.Render("press / to filter")
	}

	listHeight := m.visibleCount()
	// Reserve: 2 chars for cursor indicator + 2 for horizontal padding inside
	// the border. Anything wider would wrap the line.
	entryWidth := innerW - 4
	if entryWidth < 10 {
		entryWidth = 10
	}

	var lines []string
	end := m.offset + listHeight
	if end > len(m.entries) {
		end = len(m.entries)
	}
	for i := m.offset; i < end; i++ {
		e := m.entries[i]
		line := m.formatEntry(e, entryWidth)
		switch {
		case i == m.cursor:
			lines = append(lines, th.SidebarSelected.Render(line))
		case e.IsError:
			lines = append(lines, th.ErrorText.Render("  "+line))
		default:
			lines = append(lines, "  "+line)
		}
	}
	if len(m.entries) == 0 {
		lines = append(lines, th.MutedText.Render("  No history entries"))
	}
	// Pad list area up to listHeight so the dialog keeps a stable size.
	for len(lines) < listHeight {
		lines = append(lines, "")
	}

	countLine := fmt.Sprintf("%d entries", len(m.entries))
	if m.yankMsg != "" {
		countLine = fmt.Sprintf("%s  •  %s", countLine, m.yankMsg)
	}

	var hint string
	if m.searchMode {
		hint = th.MutedText.Render("enter/↓: list  esc: clear filter")
	} else {
		hint = th.MutedText.Render("enter: run  e: load+yank  y: yank  /: filter  j/k: nav  esc: close")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		searchLine,
		"",
		strings.Join(lines, "\n"),
		"",
		th.MutedText.Render(countLine),
		hint,
	)

	return th.FocusedBorder.
		Width(innerW).
		Padding(0, 1).
		Render(content)
}

// dialogSize returns the target outer width/height of the dialog, sized to
// mirror the floating editor overlay.
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

// visibleCount returns how many entry rows fit inside the dialog.
func (m Model) visibleCount() int {
	_, h := m.dialogSize()
	// Chrome: border(2) + title(1) + search(1) + blank(1) + blank(1) + count(1) + hint(1) = 8
	avail := h - 8
	if avail < 3 {
		avail = 3
	}
	return avail
}

func (m *Model) ensureVisible() {
	visible := m.visibleCount()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) loadEntries() {
	if m.hist == nil {
		m.entries = nil
		return
	}

	var err error
	searchText := m.search.Value()
	if searchText != "" {
		m.entries, err = m.hist.Search("%"+searchText+"%", 200)
	} else {
		m.entries, err = m.hist.Recent(200)
	}
	if err != nil {
		m.entries = nil
	}
}

func (m Model) formatEntry(e history.HistoryEntry, maxWidth int) string {
	var meta []string
	if e.Adapter != "" {
		meta = append(meta, e.Adapter)
	}
	if e.DurationMS > 0 {
		meta = append(meta, formatDuration(e.DurationMS))
	}
	meta = append(meta, RelativeTime(e.ExecutedAt))
	metaStr := strings.Join(meta, " | ")

	// Reserve space for "  " separator + meta; keep at least 10 chars for the
	// query text.
	queryMax := maxWidth - len(metaStr) - 2
	if queryMax < 10 {
		queryMax = 10
	}

	query := firstLine(e.Query)
	if len(query) > queryMax {
		query = query[:queryMax-3] + "..."
	}

	line := fmt.Sprintf("%-*s  %s", queryMax, query, metaStr)
	// Hard cap so an unusually long meta string can't overflow and wrap.
	if len(line) > maxWidth {
		line = line[:maxWidth]
	}
	return line
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// RelativeTime formats a timestamp as a human-readable relative time.
func RelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 48*time.Hour:
		return "yesterday"
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
