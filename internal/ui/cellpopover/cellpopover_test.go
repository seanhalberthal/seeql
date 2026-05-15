package cellpopover

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sized(value string) Model {
	m := New()
	m.SetSize(120, 40)
	m.Show("col", "", value)
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	}
	if len(s) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestShowMakesVisible(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	if m.Visible() {
		t.Fatal("expected popover hidden after New()")
	}
	m.Show("c", "", "hello")
	if !m.Visible() {
		t.Fatal("expected popover visible after Show()")
	}
	if m.colName != "c" || m.rawValue != "hello" {
		t.Fatalf("unexpected state: colName=%q rawValue=%q", m.colName, m.rawValue)
	}
	if m.offset != 0 {
		t.Fatalf("expected offset 0, got %d", m.offset)
	}
}

func TestHideClears(t *testing.T) {
	m := sized("hello")
	m.search.SetValue("ello")
	m.searchMode = true
	m.Hide()
	if m.Visible() {
		t.Fatal("expected hidden")
	}
	if m.searchMode {
		t.Fatal("expected searchMode cleared")
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Show("col", "", `{"a":1,"b":[2,3]}`)
	if !m.pretty {
		t.Fatalf("expected pretty=true for JSON object, got false; displayed=%q", m.displayed)
	}
	if !strings.Contains(m.displayed, "\n") {
		t.Fatal("expected multi-line pretty output")
	}
}

func TestNonJSONLeavesRaw(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Show("col", "", "just a plain string")
	if m.pretty {
		t.Fatal("expected pretty=false for non-JSON")
	}
	if m.displayed != "just a plain string" {
		t.Fatalf("displayed should equal raw, got %q", m.displayed)
	}
}

func TestMalformedJSONFallback(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Show("col", "", `{"a": 1, "b": [`) // truncated, invalid
	if m.pretty {
		t.Fatal("expected pretty=false for malformed JSON")
	}
}

func TestNavigationClamps(t *testing.T) {
	// Big content so it overflows the visible window.
	long := strings.Repeat("line\n", 200)
	m := sized(long)

	// j (down) increments offset.
	prev := m.offset
	mm, _ := m.Update(key("j"))
	if mm.offset != prev+1 {
		t.Fatalf("expected offset %d, got %d", prev+1, mm.offset)
	}

	// k (up) at offset 0 stays at 0.
	m.offset = 0
	mm, _ = m.Update(key("k"))
	if mm.offset != 0 {
		t.Fatalf("expected offset 0, got %d", mm.offset)
	}

	// G jumps to maxOffset.
	mm, _ = m.Update(key("G"))
	if mm.offset != m.maxOffset() {
		t.Fatalf("expected offset %d, got %d", m.maxOffset(), mm.offset)
	}

	// g jumps to 0.
	mm, _ = mm.Update(key("g"))
	if mm.offset != 0 {
		t.Fatalf("expected offset 0 after g, got %d", mm.offset)
	}
}

func TestPageMovementByVisibleHeight(t *testing.T) {
	long := strings.Repeat("line\n", 300)
	m := sized(long)
	visH := m.visibleLineCount()
	m.offset = 0
	mm, _ := m.Update(key("ctrl+d"))
	if mm.offset != visH {
		t.Fatalf("ctrl+d expected offset %d, got %d", visH, mm.offset)
	}
	mm, _ = mm.Update(key("ctrl+u"))
	if mm.offset != 0 {
		t.Fatalf("ctrl+u expected offset 0, got %d", mm.offset)
	}
}

func TestWrapLines(t *testing.T) {
	got := wrapLines("abcdefghij", 4)
	want := []string{"abcd", "efgh", "ij"}
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d: %#v", len(want), len(got), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("line %d: want %q got %q", i, want[i], got[i])
		}
	}
}

func TestWrapLinesBreaksAtWhitespace(t *testing.T) {
	got := wrapLines("aaa bbb ccc ddd", 7)
	want := []string{"aaa bbb", "ccc ddd"}
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d: want %q got %q", i, want[i], got[i])
		}
	}
}

func TestWrapLinesHardBreaksLongTokens(t *testing.T) {
	got := wrapLines("aaa verylongtokenwithoutspaces bbb", 8)
	for _, line := range got {
		if len(line) > 8 {
			t.Fatalf("line exceeds width: %q (%d)", line, len(line))
		}
	}
	joined := strings.Join(got, "")
	if !strings.Contains(joined, "verylongtokenwithoutspaces") {
		t.Fatalf("long token lost in wrap: %#v", got)
	}
}

func TestViewShowsColumnType(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Show("payload", "jsonb", `{"a":1}`)
	view := m.View()
	if !strings.Contains(view, "jsonb") {
		t.Fatalf("expected view to contain column type 'jsonb', got: %s", view)
	}
}

func TestWrapLinesEmptySource(t *testing.T) {
	got := wrapLines("a\n\nb", 10)
	want := []string{"a", "", "b"}
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d: %#v", len(want), len(got), got)
	}
}

func TestSearchPopulatesMatches(t *testing.T) {
	m := sized("alpha beta GAMMA alpha")
	// Type "alpha" into search via enter-and-commit path.
	m.searchMode = true
	m.search.Focus()
	m.search.SetValue("alpha")
	mm, _ := m.Update(key("enter"))
	if len(mm.matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(mm.matches))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	m := sized("Alpha BETA gamma")
	m.search.SetValue("alpha")
	m.findMatches()
	if len(m.matches) != 1 {
		t.Fatalf("expected 1 case-insensitive match, got %d", len(m.matches))
	}
}

func TestNCyclesMatches(t *testing.T) {
	long := strings.Repeat("foo\n", 50) + "bar\n" + strings.Repeat("foo\n", 50)
	m := sized(long)
	m.search.SetValue("bar")
	m.findMatches()
	if len(m.matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.matches))
	}
	// n cycles to match 0 since there's only one match.
	mm, _ := m.Update(key("n"))
	if mm.matchIdx != 0 {
		t.Fatalf("expected matchIdx 0, got %d", mm.matchIdx)
	}
}

func TestEscInSearchMode(t *testing.T) {
	m := sized("hello world")
	m.searchMode = true
	m.search.Focus()
	m.search.SetValue("world")
	m.findMatches()
	mm, _ := m.Update(key("esc"))
	if mm.searchMode {
		t.Fatal("expected searchMode cleared by esc")
	}
	if mm.search.Value() != "" {
		t.Fatalf("expected search value cleared, got %q", mm.search.Value())
	}
	if len(mm.matches) != 0 {
		t.Fatal("expected matches cleared")
	}
	// Still visible.
	if !mm.Visible() {
		t.Fatal("expected popover still visible")
	}
}

func TestEscInNavModeHides(t *testing.T) {
	m := sized("hello")
	mm, _ := m.Update(key("esc"))
	if mm.Visible() {
		t.Fatal("expected popover hidden after esc in nav mode")
	}
}

func TestQHidesInNavMode(t *testing.T) {
	m := sized("hello")
	mm, _ := m.Update(key("q"))
	if mm.Visible() {
		t.Fatal("expected q to hide popover")
	}
}

func TestEnsureMatchVisibleScrolls(t *testing.T) {
	// Construct content where a match is far below the initial viewport.
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString("filler\n")
	}
	sb.WriteString("needle\n")
	m := sized(sb.String())
	m.offset = 0
	m.search.SetValue("needle")
	m.findMatches()
	if len(m.matches) == 0 {
		t.Fatal("expected at least one match")
	}
	m.matchIdx = 0
	m.ensureMatchVisible()
	if m.offset == 0 {
		t.Fatal("expected viewport to scroll to bring match into view")
	}
}

func TestViewWhenHiddenIsEmpty(t *testing.T) {
	m := New()
	if m.View() != "" {
		t.Fatal("expected empty view when not visible")
	}
}
