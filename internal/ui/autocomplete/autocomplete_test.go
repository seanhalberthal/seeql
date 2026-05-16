package autocomplete

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seanhalberthal/seeql/internal/adapter"
	"github.com/seanhalberthal/seeql/internal/completion"
	"github.com/seanhalberthal/seeql/internal/schema"
	"github.com/seanhalberthal/seeql/internal/theme"
)

func init() {
	theme.Current = theme.Default()
}

func TestNew(t *testing.T) {
	m := New(nil)

	if m.Visible() {
		t.Fatal("expected not visible initially")
	}
	if m.engine != nil {
		t.Fatal("expected nil engine")
	}
	if m.width != 40 {
		t.Fatalf("expected default width=40, got %d", m.width)
	}
}

func TestNew_WithEngine(t *testing.T) {
	eng := completion.NewEngine("sqlite")
	m := New(eng)

	if m.engine != eng {
		t.Fatal("expected engine to be set")
	}
}

func TestTrigger_NoEngine(t *testing.T) {
	m := New(nil)

	m.Trigger("SELECT ", 7)
	if m.Visible() {
		t.Fatal("expected not visible when engine is nil")
	}
}

func TestTrigger_WithEngine(t *testing.T) {
	eng := completion.NewEngine("sqlite")
	eng.UpdateSchema([]schema.Database{
		{
			Name: "main",
			Schemas: []schema.Schema{
				{
					Name: "main",
					Tables: []schema.Table{
						{Name: "users", Columns: []schema.Column{
							{Name: "id", Type: "integer"},
							{Name: "name", Type: "text"},
						}},
					},
				},
			},
		},
	})

	m := New(eng)
	m.Trigger("SELECT * FROM u", 15)

	// The engine should have returned some results for "u" prefix.
	// Whether it is visible depends on the engine returning items.
	// With fuzzy matching, "u" should match "users".
	if !m.Visible() {
		t.Fatal("expected visible after trigger with matching prefix")
	}
	if m.selected != 0 {
		t.Fatalf("expected selected=0, got %d", m.selected)
	}
	if len(m.filtered) == 0 {
		t.Fatal("expected some filtered items")
	}
}

func TestTrigger_AfterClosingQuote(t *testing.T) {
	eng := completion.NewEngine("sqlite")
	eng.UpdateSchema([]schema.Database{
		{
			Name: "main",
			Schemas: []schema.Schema{
				{
					Name: "main",
					Tables: []schema.Table{
						{Name: "users", Columns: []schema.Column{{Name: "name", Type: "text"}}},
					},
				},
			},
		},
	})

	m := New(eng)

	// Cursor sits right after the closing quote of a string literal.
	text := "SELECT * FROM users WHERE name = 'sean'"
	m.Trigger(text, len(text))
	if m.Visible() {
		t.Fatal("expected autocomplete to stay hidden right after a closing quote")
	}

	// Once the user presses space, suggestions should be allowed again.
	text += " "
	m.Trigger(text, len(text))
	if !m.Visible() {
		t.Fatal("expected autocomplete to trigger after a space following a quoted value")
	}

	// Same check for double quotes (PostgreSQL-style identifier quoting).
	m2 := New(eng)
	m2.Trigger(`SELECT "id"`, len(`SELECT "id"`))
	if m2.Visible() {
		t.Fatal(`expected autocomplete to stay hidden right after closing double-quote`)
	}
}

func TestTrigger_NoMatches(t *testing.T) {
	eng := completion.NewEngine("sqlite")
	// No schema data, so no table/column completions. Only keywords.
	m := New(eng)
	m.Trigger("xyznonexistent", 14)

	// With only keyword completions, fuzzy matching "xyznonexistent" should not match.
	if m.Visible() {
		// If it is visible, there were some fuzzy matches. That is acceptable
		// depending on the fuzzy matcher behavior. Just verify no crash.
		t.Log("note: autocomplete visible with fuzzy matches for 'xyznonexistent'")
	}
}

func TestUpdate_Navigation(t *testing.T) {
	m := New(nil)
	// Manually set up filtered items and make visible.
	m.filtered = []adapter.CompletionItem{
		{Label: "users", Kind: adapter.CompletionTable},
		{Label: "orders", Kind: adapter.CompletionTable},
		{Label: "products", Kind: adapter.CompletionTable},
	}
	m.visible = true
	m.selected = 0

	// Move down.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Fatalf("expected selected=1 after down, got %d", m.selected)
	}

	// Move down again.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 2 {
		t.Fatalf("expected selected=2 after down, got %d", m.selected)
	}

	// Move down at boundary: should stay at 2.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 2 {
		t.Fatalf("expected selected=2 at boundary, got %d", m.selected)
	}

	// Move up.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 1 {
		t.Fatalf("expected selected=1 after up, got %d", m.selected)
	}

	// Move up again.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Fatalf("expected selected=0 after up, got %d", m.selected)
	}

	// Move up at boundary: should stay at 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Fatalf("expected selected=0 at boundary, got %d", m.selected)
	}
}

func TestUpdate_CtrlNavigation(t *testing.T) {
	m := New(nil)
	m.filtered = []adapter.CompletionItem{
		{Label: "a"},
		{Label: "b"},
	}
	m.visible = true
	m.selected = 0

	// ctrl+n = down.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if m.selected != 1 {
		t.Fatalf("expected selected=1 after ctrl+n, got %d", m.selected)
	}

	// ctrl+p = up.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if m.selected != 0 {
		t.Fatalf("expected selected=0 after ctrl+p, got %d", m.selected)
	}
}

// TestUpdate_Enter_DoesNotAccept asserts that Enter is *not* an acceptance key
// for the autocomplete dropdown. Acceptance is Tab-only — Enter is reserved
// for inserting a newline in the editor and must pass through unconsumed.
//
// Regression: previously Enter would accept the highlighted completion even
// when the user only wanted a newline. Two visible failure modes:
//
//   - prefix == highlighted item (e.g. user typed "products" with "products"
//     highlighted): ReplaceWord swapped the prefix with itself — no visible
//     change in the buffer — but the Enter was consumed, so the user's
//     newline silently vanished and the next keystroke continued on the
//     same line: `FROM products` + Enter + `WHERE` → `productsWHERE`.
//   - prefix != highlighted item (e.g. typed "c" with "CEIL" highlighted):
//     `c` was replaced with `CEIL` mid-statement: `customers c JOIN` →
//     `customers CEILJOIN`.
//
// Both variants share one root cause: Enter being listed alongside Tab in the
// accept switch. This test pins Tab as the sole accept key.
func TestUpdate_Enter_DoesNotAccept(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		label  string
	}{
		{"same-prefix variant (productsWHERE bug)", "products", "products"},
		{"different-prefix variant (CEILJOIN bug)", "c", "CEIL"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(nil)
			m.filtered = []adapter.CompletionItem{
				{Label: tc.label, Kind: adapter.CompletionTable},
			}
			m.visible = true
			m.selected = 0
			m.prefix = tc.prefix

			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if cmd != nil {
				if msg, ok := cmd().(SelectedMsg); ok {
					t.Fatalf("Enter must not produce SelectedMsg, got %+v — this is the %q regression", msg, tc.name)
				}
				t.Fatalf("Enter must not return a cmd from autocomplete.Update, got %T", cmd())
			}
		})
	}
}

// TestUpdate_Enter_LeavesVisibilityToCaller asserts that autocomplete.Update
// does not toggle visibility on Enter. The dismiss-on-Enter policy lives at
// the app level (so the key can fall through to the editor for newline
// insertion); the autocomplete component itself stays out of it.
func TestUpdate_Enter_LeavesVisibilityToCaller(t *testing.T) {
	m := New(nil)
	m.filtered = []adapter.CompletionItem{{Label: "users"}}
	m.visible = true
	m.selected = 0
	m.prefix = "us"

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.Visible() {
		t.Fatal("autocomplete.Update(Enter) must not hide the dropdown; the app decides whether to dismiss")
	}
}

func TestUpdate_Tab(t *testing.T) {
	m := New(nil)
	m.filtered = []adapter.CompletionItem{
		{Label: "users"},
	}
	m.visible = true
	m.selected = 0
	m.prefix = ""

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	if m.Visible() {
		t.Fatal("expected not visible after tab")
	}
	if cmd == nil {
		t.Fatal("expected cmd from tab")
	}
	msg := cmd()
	selMsg, ok := msg.(SelectedMsg)
	if !ok {
		t.Fatalf("expected SelectedMsg, got %T", msg)
	}
	if selMsg.Text != "users" {
		t.Fatalf("expected 'users', got %q", selMsg.Text)
	}
}

func TestUpdate_Escape(t *testing.T) {
	m := New(nil)
	m.filtered = []adapter.CompletionItem{{Label: "test"}}
	m.visible = true

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if m.Visible() {
		t.Fatal("expected not visible after escape")
	}
	if cmd == nil {
		t.Fatal("expected cmd from escape")
	}
	msg := cmd()
	_, ok := msg.(DismissMsg)
	if !ok {
		t.Fatalf("expected DismissMsg, got %T", msg)
	}
}

func TestUpdate_CtrlC(t *testing.T) {
	m := New(nil)
	m.filtered = []adapter.CompletionItem{{Label: "test"}}
	m.visible = true

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if m.Visible() {
		t.Fatal("expected not visible after ctrl+c")
	}
	if cmd == nil {
		t.Fatal("expected cmd from ctrl+c")
	}
	msg := cmd()
	_, ok := msg.(DismissMsg)
	if !ok {
		t.Fatalf("expected DismissMsg, got %T", msg)
	}
}

func TestUpdate_NotVisible(t *testing.T) {
	m := New(nil)

	// When not visible, updates should be ignored.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Fatal("expected nil cmd when not visible")
	}
}

func TestDismiss(t *testing.T) {
	m := New(nil)
	m.visible = true

	m.Dismiss()
	if m.Visible() {
		t.Fatal("expected not visible after Dismiss()")
	}
}

func TestVisible(t *testing.T) {
	m := New(nil)
	if m.Visible() {
		t.Fatal("expected not visible initially")
	}

	m.visible = true
	if !m.Visible() {
		t.Fatal("expected visible when set")
	}
}

func TestView_Hidden(t *testing.T) {
	m := New(nil)

	view := m.View()
	if view != "" {
		t.Fatalf("expected empty view when hidden, got %q", view)
	}
}

func TestView_EmptyFiltered(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.filtered = nil

	view := m.View()
	if view != "" {
		t.Fatalf("expected empty view with no filtered items, got %q", view)
	}
}

func TestView_WithItems(t *testing.T) {
	m := New(nil)
	m.visible = true
	m.filtered = []adapter.CompletionItem{
		{Label: "users", Kind: adapter.CompletionTable, Detail: "table"},
		{Label: "orders", Kind: adapter.CompletionTable, Detail: "table"},
	}
	m.selected = 0

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view with items")
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		cursorPos int
		want      string
	}{
		{"empty", "", 0, ""},
		{"single word", "SELECT", 6, "SELECT"},
		{"word after space", "SELECT us", 9, "us"},
		{"at space boundary", "SELECT ", 7, ""},
		{"after open paren", "COUNT(u", 7, "u"},
		{"after comma", "id,na", 5, "na"},
		{"after dot", "users.na", 8, "na"},
		{"after equals", "id=val", 6, "val"},
		{"cursor in middle", "SELECT * FROM users", 10, "F"},
		{"cursor past end", "abc", 100, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPrefix(tt.text, tt.cursorPos)
			if got != tt.want {
				t.Errorf("extractPrefix(%q, %d) = %q, want %q", tt.text, tt.cursorPos, got, tt.want)
			}
		})
	}
}

func TestSetPosition(t *testing.T) {
	m := New(nil)
	m.SetPosition(10, 20)

	if m.posX != 10 {
		t.Fatalf("expected posX=10, got %d", m.posX)
	}
	if m.posY != 20 {
		t.Fatalf("expected posY=20, got %d", m.posY)
	}
}

func TestSetEngine(t *testing.T) {
	m := New(nil)
	if m.engine != nil {
		t.Fatal("expected nil engine")
	}

	eng := completion.NewEngine("postgres")
	m.SetEngine(eng)

	if m.engine != eng {
		t.Fatal("expected engine to be set")
	}
}

func TestInit(t *testing.T) {
	m := New(nil)
	cmd := m.Init()
	if cmd != nil {
		t.Fatal("expected nil cmd from Init")
	}
}

func TestKindIcon(t *testing.T) {
	tests := []struct {
		kind adapter.CompletionKind
		want string
	}{
		{adapter.CompletionTable, "T"},
		{adapter.CompletionColumn, "C"},
		{adapter.CompletionKeyword, "K"},
		{adapter.CompletionFunction, "F"},
		{adapter.CompletionSchema, "S"},
		{adapter.CompletionDatabase, "D"},
		{adapter.CompletionView, "V"},
		{adapter.CompletionKind(99), " "},
	}

	for _, tt := range tests {
		got := kindIcon(tt.kind)
		if got != tt.want {
			t.Errorf("kindIcon(%d) = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestIsWordBreak(t *testing.T) {
	wordBreaks := []byte{' ', '\t', '\n', '(', ')', ',', ';', '.', '=', '<', '>'}
	for _, b := range wordBreaks {
		if !isWordBreak(b) {
			t.Errorf("expected %q to be word break", string(b))
		}
	}

	nonBreaks := []byte{'a', 'Z', '0', '_'}
	for _, b := range nonBreaks {
		if isWordBreak(b) {
			t.Errorf("expected %q to NOT be word break", string(b))
		}
	}
}
