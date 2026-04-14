package historybrowser

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seanhalberthal/seeql/internal/history"
)

func TestNilHistory(t *testing.T) {
	m := New(nil)
	m.Show()

	if !m.Visible() {
		t.Fatal("expected visible after Show()")
	}
	if len(m.entries) != 0 {
		t.Fatalf("expected 0 entries with nil history, got %d", len(m.entries))
	}

	// Should not panic on Update
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m.View()
}

func TestSelectQueryMsg(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.entries = []histEntry{
		{Query: "SELECT 1"},
		{Query: "SELECT 2"},
	}
	m.cursor = 1

	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.Visible() {
		t.Fatal("expected hidden after enter")
	}
	if cmd == nil {
		t.Fatal("expected cmd from enter")
	}
	msg := cmd()
	sel, ok := msg.(SelectQueryMsg)
	if !ok {
		t.Fatalf("expected SelectQueryMsg, got %T", msg)
	}
	if sel.Query != "SELECT 2" {
		t.Fatalf("expected 'SELECT 2', got %q", sel.Query)
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		offset time.Duration
		want   string
	}{
		{5 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{2 * time.Hour, "2h ago"},
		{36 * time.Hour, "yesterday"},
		{72 * time.Hour, "3d ago"},
	}
	for _, tt := range tests {
		got := RelativeTime(time.Now().Add(-tt.offset))
		if got != tt.want {
			t.Errorf("RelativeTime(-%v) = %q, want %q", tt.offset, got, tt.want)
		}
	}
}

func TestFormatEntryTruncation(t *testing.T) {
	m := New(nil)
	long := "SELECT very_long_column_name_one, very_long_column_name_two, very_long_column_name_three FROM some_extremely_long_table_name WHERE condition = true"
	e := histEntry{
		Query:      long,
		Adapter:    "sqlite",
		DurationMS: 42,
		ExecutedAt: time.Now().Add(-5 * time.Minute),
	}

	result := m.formatEntry(e, 60)
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	// The query part should be truncated
	if !contains(result, "...") && len(long) > 30 {
		t.Error("expected truncation with '...' for long query")
	}
}

func TestHideShow(t *testing.T) {
	m := New(nil)

	if m.Visible() {
		t.Fatal("should not be visible initially")
	}

	m.Show()
	if !m.Visible() {
		t.Fatal("should be visible after Show()")
	}

	m.Hide()
	if m.Visible() {
		t.Fatal("should not be visible after Hide()")
	}
}

func TestEscHides(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.Visible() {
		t.Fatal("esc should hide")
	}
}

// histEntry is a shorthand alias used only in tests to reduce verbosity.
type histEntry = history.HistoryEntry

func TestVimNavigation(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.entries = []histEntry{
		{Query: "SELECT 1"},
		{Query: "SELECT 2"},
		{Query: "SELECT 3"},
	}

	// j moves down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("j: expected cursor=1, got %d", m.cursor)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("jj: expected cursor=2, got %d", m.cursor)
	}
	// j at end is a no-op
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("j at end: expected cursor=2, got %d", m.cursor)
	}

	// k moves up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Fatalf("k: expected cursor=1, got %d", m.cursor)
	}

	// G jumps to bottom, g jumps to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.cursor != 2 {
		t.Fatalf("G: expected cursor=2, got %d", m.cursor)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.cursor != 0 {
		t.Fatalf("g: expected cursor=0, got %d", m.cursor)
	}
}

func TestSlashEntersSearchMode(t *testing.T) {
	m := New(nil)
	m.visible = true

	if m.searchMode {
		t.Fatal("searchMode should start false")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searchMode {
		t.Fatal("expected searchMode=true after '/'")
	}
	if !m.search.Focused() {
		t.Fatal("expected search input to be focused")
	}
}

func TestNavKeysInertInSearchMode(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.entries = []histEntry{
		{Query: "SELECT 1"},
		{Query: "SELECT 2"},
	}
	// Enter search mode.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	// 'j' should now be typed into the filter, not move the cursor.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 0 {
		t.Fatalf("j in search mode should not move cursor, got %d", m.cursor)
	}
	if got := m.search.Value(); got != "j" {
		t.Fatalf("expected filter value 'j', got %q", got)
	}

	// Esc leaves search mode and clears the filter.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.searchMode {
		t.Fatal("esc should exit search mode")
	}
	if m.search.Value() != "" {
		t.Fatal("esc should clear filter when one is set")
	}
	if !m.Visible() {
		t.Fatal("esc should not close browser when leaving a filter")
	}
}

func TestEnterExitsSearchMode(t *testing.T) {
	m := New(nil)
	m.visible = true
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.searchMode {
		t.Fatal("enter should leave search mode")
	}
	if m.search.Value() != "f" {
		t.Fatal("enter should keep the committed filter")
	}
	if !m.Visible() {
		t.Fatal("enter in search mode should not close the browser")
	}
}

func TestLoadQueryMsg(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.entries = []histEntry{{Query: "SELECT load_me"}}

	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m.Visible() {
		t.Fatal("e should close the browser")
	}
	if cmd == nil {
		t.Fatal("e should emit a command")
	}
	load, ok := cmd().(LoadQueryMsg)
	if !ok {
		t.Fatalf("expected LoadQueryMsg, got %T", cmd())
	}
	if load.Query != "SELECT load_me" {
		t.Fatalf("expected 'SELECT load_me', got %q", load.Query)
	}
}

func TestYankShowsTransientHintAndStaysOpen(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.entries = []histEntry{{Query: "SELECT yank_me"}}

	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !m.Visible() {
		t.Fatal("y should keep the browser open")
	}
	if m.yankMsg == "" {
		t.Fatal("y should set a transient yank hint")
	}
	if cmd == nil {
		t.Fatal("y should return a tea.Tick cmd to clear the hint")
	}

	// Simulate the clear tick for the current yank generation.
	gen := m.yankGen
	m, _ = m.Update(ClearYankMsg{Gen: gen})
	if m.yankMsg != "" {
		t.Fatal("ClearYankMsg should clear the hint")
	}

	// A stale ClearYankMsg should not clobber a newer yank.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m, _ = m.Update(ClearYankMsg{Gen: gen}) // stale
	if m.yankMsg == "" {
		t.Fatal("stale ClearYankMsg should not clear the newer hint")
	}
}

func TestFormatEntryFitsWithinMaxWidth(t *testing.T) {
	m := New(nil)
	cases := []struct {
		name string
		e    histEntry
	}{
		{
			name: "short meta, yesterday",
			e: histEntry{
				Query:      "SELECT * FROM cases",
				Adapter:    "postgres",
				DurationMS: 14,
				ExecutedAt: time.Now().Add(-30 * time.Hour),
			},
		},
		{
			name: "long meta",
			e: histEntry{
				Query:      "SELECT * FROM hearing_participants",
				Adapter:    "postgres",
				DurationMS: 12345,
				ExecutedAt: time.Now().Add(-30 * time.Hour),
			},
		},
		{
			name: "very long query",
			e: histEntry{
				Query:      "SELECT a_long_column, another_long_column, yet_another FROM some_very_long_named_table WHERE x = 1",
				Adapter:    "postgres",
				DurationMS: 7,
				ExecutedAt: time.Now(),
			},
		},
	}

	for _, w := range []int{40, 60, 80, 120} {
		for _, tc := range cases {
			got := m.formatEntry(tc.e, w)
			if len(got) > w {
				t.Errorf("%s @ w=%d: line length %d exceeds maxWidth", tc.name, w, len(got))
			}
			if contains(got, "\n") {
				t.Errorf("%s @ w=%d: line contains newline", tc.name, w)
			}
		}
	}
}

func TestQCloses(t *testing.T) {
	m := New(nil)
	m.Show()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.Visible() {
		t.Fatal("q should close the browser in nav mode")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
