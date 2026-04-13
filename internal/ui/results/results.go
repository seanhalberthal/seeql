// Package results provides a virtualized table component for displaying
// SQL query results. It supports both fully-loaded and streaming result
// sets with paginated fetching through adapter.RowIterator.
package results

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/atotto/clipboard"
	"github.com/mattn/go-runewidth"
	"github.com/seanhalberthal/seeql/internal/adapter"
	appmsg "github.com/seanhalberthal/seeql/internal/msg"
	"github.com/seanhalberthal/seeql/internal/theme"
)

// ClearYankMsg is sent after a timeout to clear the yank confirmation.
type ClearYankMsg struct {
	Gen   uint64
	TabID int
}

// FetchedPageMsg carries rows fetched asynchronously from an iterator.
type FetchedPageMsg struct {
	Rows    [][]string
	Forward bool // true = FetchNext, false = FetchPrev
	Err     error
	TabID   int
}

// maxBufferedRows is the maximum number of rows kept in memory for streamed
// results. When this limit is exceeded, the oldest rows are trimmed.
const maxBufferedRows = 5000

// Model is the results table component. It wraps bubbles/table with support
// for streaming large result sets via adapter.RowIterator.
type Model struct {
	table     table.Model
	columns   []adapter.ColumnMeta
	tableCols []table.Column      // computed column definitions for rendering
	rows      [][]string          // current page of rows in memory
	allRows   [][]string          // all loaded rows (for non-streaming results)
	totalRows int64               // total row count (-1 if unknown)
	offset    int                 // current scroll offset in the full dataset
	viewTop   int                 // first visible row index for custom rendering
	pageSize  int                 // rows per page
	iterator  adapter.RowIterator // for streaming results
	tabID     int
	width     int
	height    int
	focused     bool
	loading     bool
	message     string // status message ("INSERT 0 1", etc.)
	queryTime   time.Duration
	err         error
	selectedCol int // currently highlighted column for cell preview
	colOffset   int // first visible column for horizontal scrolling
	yankMsg     string
	yankGen     uint64

	// Column filter
	filtering  bool   // whether the filter text input is active
	filterText string // current filter query
	filterCol  int    // column index being filtered (-1 = none)
}

// New creates a new results model with sensible defaults.
func New(tabID int) Model {
	t := table.New(
		table.WithFocused(false),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240"))
	s.Selected = s.Selected.
		Bold(false)
	t.SetStyles(s)

	return Model{
		table:     t,
		tabID:     tabID,
		pageSize:  1000,
		totalRows: -1,
		filterCol: -1,
	}
}

// Init returns no initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the results table.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Filter input mode — capture all keys.
		if m.filtering {
			switch msg.Type {
			case tea.KeyEscape:
				m.filtering = false
				m.filterText = ""
				m.filterCol = -1
				m.applyFilter()
				m.rebuildTableRows()
				m.table.SetCursor(0)
				m.viewTop = 0
			case tea.KeyEnter:
				m.filtering = false
				// Keep filter active if there's text.
				if m.filterText == "" {
					m.filterCol = -1
				}
			case tea.KeyBackspace:
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
					m.rebuildTableRows()
					m.table.SetCursor(0)
					m.viewTop = 0
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.filterText += string(msg.Runes)
					m.applyFilter()
					m.rebuildTableRows()
					m.table.SetCursor(0)
					m.viewTop = 0
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "/":
			// Enter filter mode on the selected column.
			if len(m.columns) > 0 && len(m.allRows) > 0 {
				m.filtering = true
				m.filterText = ""
				m.filterCol = m.selectedCol
			}
			return m, nil
		case "h", "left":
			if m.selectedCol > 0 {
				m.selectedCol--
				m.ensureSelectedColVisible()
			}
			return m, nil
		case "l", "right":
			if m.selectedCol < len(m.columns)-1 {
				m.selectedCol++
				m.ensureSelectedColVisible()
			}
			return m, nil
		case "y":
			if len(m.columns) > 0 && len(m.rows) > 0 {
				val := m.selectedCellValue()
				col := m.selectedCol
				if col >= len(m.columns) {
					col = len(m.columns) - 1
				}
				colName := m.columns[col].Name
				_ = clipboard.WriteAll(val)
				m.yankMsg = fmt.Sprintf("Yanked %s to clipboard", colName)
				m.yankGen++
				gen := m.yankGen
				tabID := m.tabID
				return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
					return ClearYankMsg{Gen: gen, TabID: tabID}
				})
			}
			return m, nil
		case "pgdown":
			// If we have an iterator and are near the end of loaded rows,
			// fetch the next page.
			if m.iterator != nil && m.table.Cursor() >= len(m.rows)-1 {
				m.loading = true
				iter := m.iterator
				return m, fetchNextPage(iter, m.tabID)
			}
		case "pgup":
			// If we have an iterator and are at the top, fetch previous page.
			if m.iterator != nil && m.offset > 0 && m.table.Cursor() == 0 {
				m.loading = true
				iter := m.iterator
				return m, fetchPrevPage(iter, m.tabID)
			}
		case "esc":
			// Clear active filter when not in input mode.
			if m.filterCol >= 0 {
				m.filterText = ""
				m.filterCol = -1
				m.applyFilter()
				m.rebuildTableRows()
				m.table.SetCursor(0)
				m.viewTop = 0
				return m, nil
			}
		}

		// Delegate all other key handling to the underlying table.
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		m.updateViewTop()
		return m, cmd

	case ClearYankMsg:
		if msg.TabID == m.tabID && msg.Gen == m.yankGen {
			m.yankMsg = ""
		}
		return m, nil

	case appmsg.QueryResultMsg:
		m.SetResults(msg.Result)
		return m, nil

	case FetchedPageMsg:
		if msg.TabID != m.tabID {
			return m, nil
		}
		m.loading = false
		if msg.Err != nil {
			if !adapter.SentinelEOF(msg.Err) {
				m.err = msg.Err
			}
			return m, nil
		}
		if msg.Forward {
			firstPage := len(m.allRows) == 0
			m.allRows = append(m.allRows, msg.Rows...)
			// Trim oldest rows if exceeding buffer limit
			if len(m.allRows) > maxBufferedRows {
				excess := len(m.allRows) - maxBufferedRows
				m.allRows = m.allRows[excess:]
				m.offset += excess
			}
			m.applyFilter()
			if firstPage {
				// Recalculate column widths now that we have actual data.
				m.rebuildTable()
			} else {
				m.rebuildTableRows()
			}
		} else {
			m.allRows = append(msg.Rows, m.allRows...)
			m.offset -= len(msg.Rows)
			if m.offset < 0 {
				m.offset = 0
			}
			// Trim newest rows if exceeding buffer limit
			if len(m.allRows) > maxBufferedRows {
				m.allRows = m.allRows[:maxBufferedRows]
			}
			m.applyFilter()
			m.rebuildTableRows()
		}
		return m, nil
	}

	// Pass through any other messages to the table.
	if m.focused {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the results component.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	th := theme.Current

	// Loading state.
	if m.loading && len(m.rows) == 0 {
		msg := th.MutedText.Render("  Executing query...")
		return m.wrapBorder(msg)
	}

	// Error state.
	if m.err != nil {
		errText := th.ErrorText.Render("  Error: " + m.err.Error())
		return m.wrapBorder(errText)
	}

	// Non-SELECT result message (INSERT, UPDATE, CREATE TABLE, etc.).
	if m.message != "" && len(m.rows) == 0 {
		msgText := th.SuccessText.Render("  " + m.message)
		return m.wrapBorder(msgText)
	}

	// Empty result set.
	if len(m.columns) == 0 && len(m.rows) == 0 && m.message == "" {
		placeholder := th.MutedText.Render("  No results — write a query and press F5 to execute")
		return m.wrapBorder(placeholder)
	}

	// Render table with custom zebra striping.
	tableView := m.renderTable()

	// Build footer.
	footer := m.buildFooter()

	content := lipgloss.JoinVertical(lipgloss.Left, tableView, footer)
	return m.wrapBorder(content)
}

// SetResults loads a complete QueryResult into the table.
func (m *Model) SetResults(result *adapter.QueryResult) {
	m.err = nil
	m.loading = false
	if m.iterator != nil {
		m.iterator.Close()
		m.iterator = nil
	}
	m.offset = 0
	m.queryTime = result.Duration

	if !result.IsSelect {
		// Non-SELECT statement: show message only.
		m.message = result.Message
		m.columns = nil
		m.rows = nil
		m.allRows = nil
		m.totalRows = result.RowCount
		m.table.SetRows(nil)
		m.table.SetColumns(nil)
		return
	}

	m.message = ""
	m.columns = result.Columns
	m.allRows = result.Rows
	m.rows = result.Rows
	m.totalRows = result.RowCount
	m.viewTop = 0
	m.selectedCol = 0
	m.filtering = false
	m.filterText = ""
	m.filterCol = -1
	if m.totalRows < 0 {
		m.totalRows = int64(len(result.Rows))
	}

	m.rebuildTable()
}

// SetIterator configures the model for streaming mode with the given iterator.
func (m *Model) SetIterator(iter adapter.RowIterator) {
	if m.iterator != nil {
		m.iterator.Close()
	}
	m.iterator = iter
	m.columns = iter.Columns()
	m.totalRows = iter.TotalRows()
	m.offset = 0
	m.viewTop = 0
	m.colOffset = 0
	m.err = nil
	m.message = ""
	m.allRows = nil
	m.rows = nil

	// Build column headers immediately so the table structure is visible.
	m.tableCols = autoSizeColumns(m.columns, nil, m.contentWidth())
	m.table.SetColumns(m.tableCols)
	m.table.SetRows(nil)
}

// SetSize updates the component dimensions and recalculates table layout.
func (m *Model) SetSize(w, h int) {
	dimChanged := m.width != w || m.height != h
	m.width = w
	m.height = h

	// Account for border + actual footer height.
	innerW := w - 2
	if innerW < 0 {
		innerW = 0
	}
	innerH := h - 2 - m.footerLineCount() // border top/bottom + footer lines
	if innerH < 1 {
		innerH = 1
	}

	m.table.SetWidth(innerW)
	m.table.SetHeight(innerH)

	// Recalculate column widths if dimensions actually changed.
	if dimChanged && len(m.columns) > 0 {
		m.tableCols = autoSizeColumns(m.columns, m.rows, m.contentWidth())
		m.table.SetColumns(m.tableCols)
	}
}

// SetLoading sets the loading state.
func (m *Model) SetLoading(loading bool) {
	m.loading = loading
	if loading {
		m.err = nil
	}
}

// SetError sets the error state.
func (m *Model) SetError(err error) {
	m.err = err
	m.loading = false
}

// SetMessage sets a status message with the associated query duration.
func (m *Model) SetMessage(msg string, duration time.Duration) {
	m.message = msg
	m.queryTime = duration
	m.err = nil
	m.loading = false
}

// Focus gives the results table keyboard focus.
func (m *Model) Focus() {
	m.focused = true
	m.table.Focus()
	m.applyStyles()
}

// Blur removes keyboard focus from the results table.
func (m *Model) Blur() {
	m.focused = false
	m.table.Blur()
	m.applyStyles()
}

// Focused reports whether the results table is currently focused.
func (m Model) Focused() bool {
	return m.focused
}

// SelectedCol returns the index of the currently highlighted column.
func (m Model) SelectedCol() int {
	return m.selectedCol
}

// selectedCellValue returns the value of the currently highlighted cell.
func (m Model) selectedCellValue() string {
	cursor := m.table.Cursor()
	col := m.selectedCol
	if col >= len(m.columns) {
		col = len(m.columns) - 1
	}
	if cursor >= 0 && cursor < len(m.rows) && col >= 0 && col < len(m.rows[cursor]) {
		return m.rows[cursor][col]
	}
	return ""
}

// SelectedRow returns the data for the currently selected row, or nil if
// no row is selected.
func (m Model) SelectedRow() []string {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return nil
	}
	return row
}

// RowCount returns the total number of rows in the result set. Returns -1
// if the total is unknown (streaming mode before completion).
func (m Model) RowCount() int64 {
	return m.totalRows
}

// QueryDuration returns how long the query took to execute.
func (m Model) QueryDuration() time.Duration {
	return m.queryTime
}

// Columns returns the current column metadata.
func (m Model) Columns() []adapter.ColumnMeta {
	return m.columns
}

// Filtering returns true when the filter text input is active and should
// capture all key input.
func (m Model) Filtering() bool {
	return m.filtering
}

// Rows returns all loaded rows.
func (m Model) Rows() [][]string {
	return m.allRows
}

// CloseIterator closes the current iterator if any, releasing resources.
func (m *Model) CloseIterator() {
	if m.iterator != nil {
		m.iterator.Close()
		m.iterator = nil
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// rebuildTable recalculates columns and repopulates the table widget.
func (m *Model) rebuildTable() {
	m.colOffset = 0
	m.tableCols = autoSizeColumns(m.columns, m.rows, m.contentWidth())
	m.table.SetColumns(m.tableCols)
	m.rebuildTableRows()
	m.table.GotoTop()
}

// rebuildTableRows converts [][]string rows into table.Row and sets them.
func (m *Model) rebuildTableRows() {
	tableRows := make([]table.Row, len(m.rows))
	for i, row := range m.rows {
		tableRows[i] = table.Row(row)
	}
	m.table.SetRows(tableRows)
}

// applyFilter filters m.allRows into m.rows based on the active filter.
// If no filter is set, m.rows is set to m.allRows.
func (m *Model) applyFilter() {
	if m.filterCol < 0 || m.filterText == "" {
		m.rows = m.allRows
		return
	}

	col := m.filterCol
	needle := strings.ToLower(m.filterText)
	var filtered [][]string
	for _, row := range m.allRows {
		if col < len(row) && strings.Contains(strings.ToLower(row[col]), needle) {
			filtered = append(filtered, row)
		}
	}
	m.rows = filtered
}

// contentWidth returns the usable width inside the border.
func (m *Model) contentWidth() int {
	w := m.width - 2 // border left + right
	if w < 10 {
		w = 10
	}
	return w
}

// footerLineCount returns the number of lines the footer will render.
func (m Model) footerLineCount() int {
	n := 0
	if m.filtering || m.filterCol >= 0 {
		n++ // filter input / active filter
	}
	if len(m.columns) > 0 && len(m.rows) > 0 && m.focused {
		n++ // cell preview
	}
	if m.yankMsg != "" {
		n++ // yank confirmation
	}
	if m.totalRows >= 0 || len(m.allRows) > 0 || m.queryTime > 0 || m.loading {
		n++ // stats line
	}
	if n == 0 {
		n = 1 // reserve at least 1 line for footer spacing
	}
	return n
}

// visCol describes a column visible in the current scroll position with its
// effective display width (which may be wider than the natural width when
// extra space is distributed among visible columns).
type visCol struct {
	idx   int // index into m.tableCols / m.columns / row data
	width int // effective content width (excluding padding)
}

// visibleColumns returns the columns that fit on screen starting from
// colOffset, distributing any leftover horizontal space among them.
func (m Model) visibleColumns() []visCol {
	if len(m.tableCols) == 0 {
		return nil
	}

	maxW := m.contentWidth()
	start := m.colOffset
	if start >= len(m.tableCols) {
		start = 0
	}

	// Collect columns that fit.
	var cols []visCol
	used := 0
	for i := start; i < len(m.tableCols); i++ {
		colW := m.tableCols[i].Width + 2 // content + padding
		if used+colW > maxW && len(cols) > 0 {
			break
		}
		cols = append(cols, visCol{idx: i, width: m.tableCols[i].Width})
		used += colW
	}

	// Distribute leftover space evenly.
	extra := maxW - used
	if extra > 0 && len(cols) > 0 {
		perCol := extra / len(cols)
		remainder := extra % len(cols)
		for i := range cols {
			cols[i].width += perCol
			if i < remainder {
				cols[i].width++
			}
		}
	}

	return cols
}

// ensureSelectedColVisible adjusts colOffset so that selectedCol is within
// the visible column range.
func (m *Model) ensureSelectedColVisible() {
	if len(m.tableCols) == 0 {
		return
	}
	if m.colOffset >= len(m.tableCols) {
		m.colOffset = 0
	}

	// Scrolled past the left edge.
	if m.selectedCol < m.colOffset {
		m.colOffset = m.selectedCol
		return
	}

	// Check whether selectedCol is already visible.
	visCols := m.visibleColumns()
	if len(visCols) > 0 && m.selectedCol <= visCols[len(visCols)-1].idx {
		return
	}

	// Scroll right: place selectedCol as the last visible column.
	maxW := m.contentWidth()
	used := m.tableCols[m.selectedCol].Width + 2
	newOffset := m.selectedCol
	for i := m.selectedCol - 1; i >= 0; i-- {
		colW := m.tableCols[i].Width + 2
		if used+colW > maxW {
			break
		}
		used += colW
		newOffset = i
	}
	m.colOffset = newOffset
}

// visibleDataHeight returns the number of data rows that can be displayed,
// accounting for the header row (1 line) and its bottom border (1 line).
func (m Model) visibleDataHeight() int {
	innerH := m.height - 2 - m.footerLineCount() // border top/bottom + footer
	h := innerH - 2                               // header + border line
	if h < 1 {
		h = 1
	}
	return h
}

// updateViewTop adjusts the scroll offset so the cursor remains visible.
func (m *Model) updateViewTop() {
	cursor := m.table.Cursor()
	visH := m.visibleDataHeight()
	if cursor < m.viewTop {
		m.viewTop = cursor
	}
	if cursor >= m.viewTop+visH {
		m.viewTop = cursor - visH + 1
	}
	if m.viewTop < 0 {
		m.viewTop = 0
	}
}

// renderTable produces the custom table view with zebra-striped rows.
func (m Model) renderTable() string {
	if len(m.tableCols) == 0 {
		return ""
	}

	th := theme.Current
	contentW := m.contentWidth()
	visH := m.visibleDataHeight()
	visCols := m.visibleColumns()

	var sb strings.Builder

	// Header row.
	sb.WriteString(m.renderHeader(th, contentW, visCols))
	sb.WriteByte('\n')

	// Header bottom border.
	sb.WriteString(strings.Repeat("─", contentW))
	sb.WriteByte('\n')

	// Data rows.
	cursor := m.table.Cursor()
	nRows := len(m.rows)
	for i := 0; i < visH; i++ {
		rowIdx := m.viewTop + i
		if rowIdx >= nRows {
			// Pad remaining lines so the table height stays constant.
			sb.WriteString(strings.Repeat(" ", contentW))
		} else {
			sb.WriteString(m.renderDataRow(th, rowIdx, rowIdx == cursor, contentW, visCols))
		}
		if i < visH-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// renderHeader renders the column header row using only the visible columns.
func (m Model) renderHeader(th *theme.Theme, totalWidth int, visCols []visCol) string {
	var sb strings.Builder
	used := 0
	for _, vc := range visCols {
		cellWidth := vc.width + 2 // +2 for Padding(0,1)
		text := runewidth.Truncate(m.tableCols[vc.idx].Title, vc.width, "…")
		text = padRight(text, vc.width)
		style := th.ResultsHeader
		if m.focused && vc.idx == m.selectedCol {
			style = th.ResultsHeaderSelected
		}
		sb.WriteString(style.Render(text))
		used += cellWidth
	}
	// Pad remainder so the header background fills the full width.
	if used < totalWidth {
		sb.WriteString(th.ResultsHeader.Padding(0).Render(strings.Repeat(" ", totalWidth-used)))
	}
	return sb.String()
}

// renderDataRow renders a single data row with zebra striping using only
// the visible columns.
func (m Model) renderDataRow(th *theme.Theme, rowIdx int, selected bool, totalWidth int, visCols []visCol) string {
	var cellStyle lipgloss.Style
	switch {
	case selected:
		cellStyle = th.ResultsSelectedRow
	case rowIdx%2 == 1:
		cellStyle = th.ResultsCellAlt
	default:
		cellStyle = th.ResultsCell
	}

	row := m.rows[rowIdx]
	var sb strings.Builder
	used := 0
	for _, vc := range visCols {
		cellWidth := vc.width + 2 // +2 for Padding(0,1)
		var val string
		if vc.idx < len(row) {
			val = row[vc.idx]
		}
		text := runewidth.Truncate(val, vc.width, "…")
		text = padRight(text, vc.width)
		rendered := cellStyle.Render(text)
		sb.WriteString(rendered)
		used += cellWidth
	}
	// Fill remaining width so the row background extends to the edge.
	if used < totalWidth {
		sb.WriteString(cellStyle.Padding(0).Render(strings.Repeat(" ", totalWidth-used)))
	}
	return sb.String()
}

// padRight pads s with spaces on the right so its display width equals w.
func padRight(s string, w int) string {
	sw := runewidth.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

// buildFooter constructs the cell preview and row count footer.
func (m Model) buildFooter() string {
	th := theme.Current
	var lines []string

	// Filter input or active filter indicator.
	if m.filtering && m.filterCol >= 0 && m.filterCol < len(m.columns) {
		colName := m.columns[m.filterCol].Name
		filterLine := fmt.Sprintf("  / Filter (%s): %s█", colName, m.filterText)
		lines = append(lines, th.MutedText.Render(filterLine))
	} else if m.filterCol >= 0 && m.filterCol < len(m.columns) && m.filterText != "" {
		colName := m.columns[m.filterCol].Name
		matchCount := len(m.rows)
		filterLine := fmt.Sprintf("  Filter (%s): %s (%d matches)", colName, m.filterText, matchCount)
		lines = append(lines, th.MutedText.Render(filterLine))
	}

	// Cell preview — show the full value of the selected column for the
	// current row so truncated values can be read.
	if len(m.columns) > 0 && len(m.rows) > 0 && m.focused {
		cursor := m.table.Cursor()
		col := m.selectedCol
		if col >= len(m.columns) {
			col = len(m.columns) - 1
		}
		colName := m.columns[col].Name
		val := ""
		if cursor >= 0 && cursor < len(m.rows) && col < len(m.rows[cursor]) {
			val = m.rows[cursor][col]
		}
		if val == "" {
			val = "NULL"
		}
		maxW := m.contentWidth()
		preview := fmt.Sprintf("  %s: %s", colName, val)
		if runewidth.StringWidth(preview) > maxW && maxW > 4 {
			preview = runewidth.Truncate(preview, maxW, "...")
		}
		lines = append(lines, th.MutedText.Render(preview))
	}

	// Yank confirmation.
	if m.yankMsg != "" {
		lines = append(lines, th.SuccessText.Render("  "+m.yankMsg))
	}

	// Stats line.
	var parts []string
	switch {
	case m.totalRows >= 0:
		parts = append(parts, fmt.Sprintf("%d rows", m.totalRows))
	case len(m.allRows) > 0:
		parts = append(parts, fmt.Sprintf("%d rows loaded", len(m.allRows)))
	}
	if m.queryTime > 0 {
		parts = append(parts, formatDuration(m.queryTime))
	}
	if m.loading {
		parts = append(parts, "loading...")
	}
	// Column scroll indicator when not all columns fit on screen.
	if visCols := m.visibleColumns(); len(m.tableCols) > 0 && len(visCols) < len(m.tableCols) {
		first := visCols[0].idx + 1
		last := visCols[len(visCols)-1].idx + 1
		parts = append(parts, fmt.Sprintf("cols %d–%d of %d", first, last, len(m.tableCols)))
	}
	if len(parts) > 0 {
		lines = append(lines, th.MutedText.Render("  "+strings.Join(parts, " | ")))
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// wrapBorder renders the content inside a themed border frame that fills
// the full allocated height.
func (m Model) wrapBorder(content string) string {
	th := theme.Current

	var borderStyle lipgloss.Style
	if m.focused {
		borderStyle = th.FocusedBorder
	} else {
		borderStyle = th.UnfocusedBorder
	}

	innerW := m.width - 2
	if innerW < 0 {
		innerW = 0
	}

	innerH := m.height - 2 // border top + bottom
	if innerH < 1 {
		innerH = 1
	}

	return borderStyle.Width(innerW).Height(innerH).Render(content)
}

// applyStyles updates the table styles based on the current theme and focus.
func (m *Model) applyStyles() {
	th := theme.Current
	s := table.DefaultStyles()

	s.Header = th.ResultsHeader.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240"))

	s.Cell = th.ResultsCell

	s.Selected = th.ResultsSelectedRow

	m.table.SetStyles(s)
}

// formatDuration produces a human-readable duration string.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%d us", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.2f s", d.Seconds())
	default:
		return fmt.Sprintf("%.1f min", d.Minutes())
	}
}

// SetQueryDuration sets the query execution time for the footer display.
func (m *Model) SetQueryDuration(d time.Duration) {
	m.queryTime = d
}

// FetchFirstPage returns a tea.Cmd that fetches the first page from an iterator.
func FetchFirstPage(iter adapter.RowIterator, tabID int) tea.Cmd {
	return fetchNextPage(iter, tabID)
}

// ---------------------------------------------------------------------------
// Async fetch commands
// ---------------------------------------------------------------------------

// fetchNextPage returns a tea.Cmd that fetches the next page from an iterator.
func fetchNextPage(iter adapter.RowIterator, tabID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rows, err := iter.FetchNext(ctx)
		return FetchedPageMsg{Rows: rows, Forward: true, Err: err, TabID: tabID}
	}
}

// fetchPrevPage returns a tea.Cmd that fetches the previous page from an iterator.
func fetchPrevPage(iter adapter.RowIterator, tabID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rows, err := iter.FetchPrev(ctx)
		return FetchedPageMsg{Rows: rows, Forward: false, Err: err, TabID: tabID}
	}
}

// ---------------------------------------------------------------------------
// Column auto-sizing
// ---------------------------------------------------------------------------

// autoSizeColumns calculates natural column widths based on header names and
// data content, capping individual columns at 50 characters. Horizontal
// scrolling handles overflow when columns exceed available width.
func autoSizeColumns(cols []adapter.ColumnMeta, rows [][]string, _ int) []table.Column {
	if len(cols) == 0 {
		return nil
	}

	numCols := len(cols)

	// Start with header lengths as minimum widths.
	const minColWidth = 10
	widths := make([]int, numCols)
	for i, c := range cols {
		widths[i] = len(c.Name)
		if widths[i] < minColWidth {
			widths[i] = minColWidth
		}
	}

	// Sample up to 100 rows to estimate content widths.
	sampleSize := len(rows)
	if sampleSize > 100 {
		sampleSize = 100
	}
	for i := 0; i < sampleSize; i++ {
		for j := 0; j < numCols && j < len(rows[i]); j++ {
			cellLen := len(rows[i][j])
			if cellLen > widths[j] {
				widths[j] = cellLen
			}
		}
	}

	// Cap individual column widths at 50 characters.
	const maxColWidth = 50
	for i := range widths {
		if widths[i] > maxColWidth {
			widths[i] = maxColWidth
		}
	}

	// No proportional scaling — horizontal scrolling handles overflow.

	// Build table.Column slice.
	tableCols := make([]table.Column, numCols)
	for i, c := range cols {
		title := c.Name
		tableCols[i] = table.Column{
			Title: title,
			Width: widths[i],
		}
	}

	return tableCols
}
