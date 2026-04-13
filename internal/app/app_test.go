package app

import (
	"context"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seanhalberthal/seeql/internal/adapter"
	"github.com/seanhalberthal/seeql/internal/config"
	"github.com/seanhalberthal/seeql/internal/schema"
)

// ---------------------------------------------------------------------------
// TestNew: default config
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	t.Run("focusedPane is PaneSidebar", func(t *testing.T) {
		if m.focusedPane != PaneSidebar {
			t.Errorf("focusedPane = %d, want PaneSidebar (%d)", m.focusedPane, PaneSidebar)
		}
	})

	t.Run("showSidebar is true", func(t *testing.T) {
		if !m.showSidebar {
			t.Error("showSidebar should be true by default")
		}
	})

	t.Run("tabStates has one entry for tab 0", func(t *testing.T) {
		if len(m.tabStates) != 1 {
			t.Fatalf("tabStates length = %d, want 1", len(m.tabStates))
		}
		ts, ok := m.tabStates[0]
		if !ok {
			t.Fatal("tabStates[0] does not exist")
		}
		if ts == nil {
			t.Fatal("tabStates[0] is nil")
		}
	})

	t.Run("sidebarWidth has default", func(t *testing.T) {
		if m.sidebarWidth != 30 {
			t.Errorf("sidebarWidth = %d, want 30", m.sidebarWidth)
		}
	})

	t.Run("showEditor is false by default", func(t *testing.T) {
		if m.showEditor {
			t.Error("showEditor should be false by default")
		}
	})

	t.Run("conn is nil", func(t *testing.T) {
		if m.conn != nil {
			t.Error("conn should be nil initially")
		}
	})

	t.Run("not quitting", func(t *testing.T) {
		if m.quitting {
			t.Error("quitting should be false initially")
		}
	})

	t.Run("not executing", func(t *testing.T) {
		if m.executing {
			t.Error("executing should be false initially")
		}
	})

	t.Run("showHelp is false", func(t *testing.T) {
		if m.showHelp {
			t.Error("showHelp should be false initially")
		}
	})

	t.Run("config is stored", func(t *testing.T) {
		if m.cfg != cfg {
			t.Error("cfg pointer does not match the config passed to New")
		}
	})

	t.Run("history is nil when passed nil", func(t *testing.T) {
		if m.history != nil {
			t.Error("history should be nil when nil was passed")
		}
	})

	t.Run("compEngine is not nil", func(t *testing.T) {
		if m.compEngine == nil {
			t.Error("compEngine should not be nil after New")
		}
	})

	t.Run("standard keymap is used", func(t *testing.T) {
		if !containsKey(m.keyMap.Quit, "ctrl+q") {
			t.Error("keyMap.Quit should contain ctrl+q")
		}
		if !containsKey(m.keyMap.ExecuteQuery, "f5") {
			t.Error("keyMap.ExecuteQuery should contain f5")
		}
	})
}

func TestNew_DefaultsUnchanged(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	t.Run("other defaults unchanged", func(t *testing.T) {
		if m.focusedPane != PaneSidebar {
			t.Errorf("focusedPane = %d, want PaneSidebar", m.focusedPane)
		}
		if !m.showSidebar {
			t.Error("showSidebar should be true by default")
		}
		if len(m.tabStates) != 1 {
			t.Errorf("tabStates length = %d, want 1", len(m.tabStates))
		}
	})
}

// ---------------------------------------------------------------------------
// TestNew_WithConnections
// ---------------------------------------------------------------------------

func TestNew_WithConnections(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Connections = []config.SavedConnection{
		{Name: "prod-pg", DSN: "postgres://localhost:5432/proddb"},
		{Name: "local-sqlite", DSN: "/tmp/test.db"},
	}
	m := New(cfg, nil, nil)

	// Should still create normally without crashing
	if m.cfg != cfg {
		t.Error("cfg should be stored in model")
	}
	if len(m.cfg.Connections) != 2 {
		t.Errorf("cfg.Connections length = %d, want 2", len(m.cfg.Connections))
	}
}

// ---------------------------------------------------------------------------
// TestIsTypingKey
// ---------------------------------------------------------------------------

func TestIsTypingKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		// Regular printable characters
		{"letter a", "a", true},
		{"letter z", "z", true},
		{"letter A", "A", true},
		{"letter Z", "Z", true},
		{"digit 0", "0", true},
		{"digit 9", "9", true},
		{"space", " ", true},
		{"exclamation", "!", true},
		{"at sign", "@", true},
		{"hash", "#", true},
		{"dollar", "$", true},
		{"percent", "%", true},
		{"ampersand", "&", true},
		{"asterisk", "*", true},
		{"open paren", "(", true},
		{"close paren", ")", true},
		{"semicolon", ";", true},
		{"dot", ".", true},
		{"comma", ",", true},
		{"equals", "=", true},
		{"plus", "+", true},
		{"minus", "-", true},
		{"underscore", "_", true},
		{"tilde", "~", true},

		// Backspace and delete are typing keys
		{"backspace", "backspace", true},
		{"delete", "delete", true},

		// Non-typing keys
		{"ctrl+c", "ctrl+c", false},
		{"enter", "enter", false},
		{"tab", "tab", false},
		{"esc", "esc", false},
		{"up", "up", false},
		{"down", "down", false},
		{"left", "left", false},
		{"right", "right", false},
		{"ctrl+q", "ctrl+q", false},
		{"ctrl+t", "ctrl+t", false},
		{"f1", "f1", false},
		{"f5", "f5", false},
		{"shift+tab", "shift+tab", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := keyMsgFromString(tt.key)
			got := isTypingKey(msg)
			if got != tt.want {
				t.Errorf("isTypingKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestInit
// ---------------------------------------------------------------------------

func TestInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil (no background tasks)")
	}
}

// ---------------------------------------------------------------------------
// TestView_BeforeWindowSize
// ---------------------------------------------------------------------------

func TestView_BeforeWindowSize(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)
	// Before receiving a WindowSizeMsg, width and height are 0
	view := m.View()
	if view != "Loading..." {
		t.Errorf("View() before WindowSize = %q, want %q", view, "Loading...")
	}
}

// ---------------------------------------------------------------------------
// TestView_Quitting
// ---------------------------------------------------------------------------

func TestView_Quitting(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)
	m.quitting = true
	view := m.View()
	if view != "Goodbye!\n" {
		t.Errorf("View() when quitting = %q, want %q", view, "Goodbye!\n")
	}
}

func TestUpdate_SwitchTabMsg_BlursInactiveTabs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	// Create a second tab.
	model, cmd := m.Update(NewTabMsg{})
	m = model.(Model)
	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}

	// Switch to tab 0 — focus should stay on sidebar (the default pane).
	model, _ = m.Update(SwitchTabMsg{TabID: 0})
	m = model.(Model)
	if !m.sidebar.Focused() {
		t.Fatal("expected sidebar focused after switching tabs")
	}

	// Simulate Ctrl+] path where tab model advances active tab before message delivery.
	nextCmd := m.tabs.NextTab()
	if nextCmd == nil {
		t.Fatal("expected NextTab command")
	}
	model, _ = m.Update(nextCmd())
	m = model.(Model)

	if m.tabs.ActiveID() != 1 {
		t.Fatalf("expected active tab ID 1, got %d", m.tabs.ActiveID())
	}
	// Sidebar stays focused across tab switches; inactive tab results should be blurred.
	if !m.sidebar.Focused() {
		t.Fatal("expected sidebar focused after tab switch")
	}
	if ts := m.tabStates[0]; ts != nil && ts.Results.Focused() {
		t.Fatal("expected inactive tab 0 results blurred after switch")
	}
}

type testConn struct {
	dbName      string
	cancelCalls int
	closed      bool
}

func (c *testConn) Databases(context.Context) ([]schema.Database, error) {
	return nil, nil
}
func (c *testConn) Tables(context.Context, string, string) ([]schema.Table, error) {
	return nil, nil
}
func (c *testConn) Columns(context.Context, string, string, string) ([]schema.Column, error) {
	return nil, nil
}
func (c *testConn) Indexes(context.Context, string, string, string) ([]schema.Index, error) {
	return nil, nil
}
func (c *testConn) ForeignKeys(context.Context, string, string, string) ([]schema.ForeignKey, error) {
	return nil, nil
}
func (c *testConn) Execute(context.Context, string) (*adapter.QueryResult, error) {
	return nil, nil
}
func (c *testConn) Cancel() error {
	c.cancelCalls++
	return nil
}
func (c *testConn) ExecuteStreaming(context.Context, string, int) (adapter.RowIterator, error) {
	return nil, nil
}
func (c *testConn) Completions(context.Context) ([]adapter.CompletionItem, error) {
	return nil, nil
}
func (c *testConn) Ping(context.Context) error { return nil }
func (c *testConn) Close() error {
	c.closed = true
	return nil
}
func (c *testConn) DatabaseName() string { return c.dbName }
func (c *testConn) AdapterName() string  { return "test" }

type testIter struct {
	closed bool
}

func (it *testIter) FetchNext(context.Context) ([][]string, error) { return nil, io.EOF }
func (it *testIter) FetchPrev(context.Context) ([][]string, error) { return nil, io.EOF }
func (it *testIter) Columns() []adapter.ColumnMeta                 { return []adapter.ColumnMeta{{Name: "col"}} }
func (it *testIter) TotalRows() int64                              { return -1 }
func (it *testIter) Close() error {
	it.closed = true
	return nil
}

func TestUpdate_ConnectMsg_CleansPreviousConnectionState(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	oldConn := &testConn{dbName: "old"}
	newConn := &testConn{dbName: "new"}
	m.conn = oldConn
	m.executing = true

	cancelCalled := false
	m.cancelFunc = func() { cancelCalled = true }

	iter := &testIter{}
	ts := m.tabStates[0]
	ts.Results.SetIterator(iter)

	model, _ := m.Update(ConnectMsg{
		Conn:    newConn,
		Adapter: "test",
		DSN:     "test://new",
	})
	m = model.(Model)

	if !cancelCalled {
		t.Fatal("expected previous query context to be cancelled on reconnect")
	}
	if oldConn.cancelCalls != 1 {
		t.Fatalf("expected old connection cancel to be called once, got %d", oldConn.cancelCalls)
	}
	if !oldConn.closed {
		t.Fatal("expected old connection to be closed on reconnect")
	}
	if !iter.closed {
		t.Fatal("expected existing iterator to be closed on reconnect")
	}
	if m.conn != newConn {
		t.Fatal("expected new connection to be installed")
	}
	if m.executing {
		t.Fatal("expected executing=false after reconnect")
	}
}

// ---------------------------------------------------------------------------
// TestCollapsedSidebarNotFocusable
// ---------------------------------------------------------------------------

func TestCollapsedSidebarNotFocusable(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	// Give it a size so layout works.
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(Model)

	t.Run("toggle sidebar moves focus to results", func(t *testing.T) {
		// Start with sidebar focused.
		m.setFocus(PaneSidebar)
		if m.focusedPane != PaneSidebar {
			t.Fatal("precondition: expected sidebar focused")
		}

		// Hide sidebar via ctrl+s.
		model, _ = m.Update(keyMsgFromString("ctrl+s"))
		m = model.(Model)

		if m.showSidebar {
			t.Fatal("expected sidebar hidden after ctrl+s")
		}
		if m.focusedPane != PaneResults {
			t.Errorf("focusedPane = %d, want PaneResults (%d)", m.focusedPane, PaneResults)
		}
	})

	t.Run("tab does not cycle to sidebar when hidden", func(t *testing.T) {
		// Sidebar is already hidden from previous subtest.
		m.setFocus(PaneResults)
		m.cycleFocus(1)
		if m.focusedPane != PaneResults {
			t.Errorf("focusedPane = %d after tab cycle, want PaneResults", m.focusedPane)
		}
	})

	t.Run("alt+1 does not focus sidebar when hidden", func(t *testing.T) {
		// Sidebar is still hidden.
		m.setFocus(PaneResults)
		model, _ = m.Update(keyMsgFromString("alt+1"))
		m = model.(Model)
		if m.focusedPane != PaneResults {
			t.Errorf("focusedPane = %d after alt+1, want PaneResults", m.focusedPane)
		}
	})

	t.Run("re-show sidebar allows focus again", func(t *testing.T) {
		// Show sidebar again.
		model, _ = m.Update(keyMsgFromString("ctrl+s"))
		m = model.(Model)
		if !m.showSidebar {
			t.Fatal("expected sidebar visible after second ctrl+s")
		}

		// Now alt+1 should work.
		model, _ = m.Update(keyMsgFromString("alt+1"))
		m = model.(Model)
		if m.focusedPane != PaneSidebar {
			t.Errorf("focusedPane = %d after alt+1, want PaneSidebar", m.focusedPane)
		}
	})
}

// ---------------------------------------------------------------------------
// keyMsgFromString creates a tea.KeyMsg from a string representation.
// This handles common key names by mapping to the appropriate KeyType.
// ---------------------------------------------------------------------------

func keyMsgFromString(s string) tea.KeyMsg {
	// For single printable characters (length 1, ASCII 32-126)
	if len(s) == 1 && s[0] >= 32 && s[0] <= 126 {
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(s),
		}
	}

	// Map named keys to their bubbletea KeyType
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+q":
		return tea.KeyMsg{Type: tea.KeyCtrlQ}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "ctrl+t":
		return tea.KeyMsg{Type: tea.KeyCtrlT}
	case "alt+1":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}, Alt: true}
	case "alt+2":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}, Alt: true}
	case "alt+3":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}, Alt: true}
	case "f1":
		return tea.KeyMsg{Type: tea.KeyF1}
	case "f5":
		return tea.KeyMsg{Type: tea.KeyF5}
	default:
		// Fallback: treat as runes (multi-byte string)
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(s),
		}
	}
}

// ---------------------------------------------------------------------------
// TestBracketKeysSwitchTabs: pressing ] switches to next tab via Update()
// ---------------------------------------------------------------------------

func TestBracketKeysSwitchTabs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	// Give it a window size so layout works.
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(Model)

	// Create a second tab.
	model, cmd := m.Update(NewTabMsg{})
	m = model.(Model)
	// Process the SwitchTabMsg from NewTab.
	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}
	if m.tabs.ActiveID() != 1 {
		t.Fatalf("precondition: expected active tab 1 after creating second tab, got %d", m.tabs.ActiveID())
	}

	// Switch to tab 0 first.
	model, _ = m.Update(SwitchTabMsg{TabID: 0})
	m = model.(Model)
	if m.tabs.ActiveID() != 0 {
		t.Fatalf("precondition: expected active tab 0, got %d", m.tabs.ActiveID())
	}

	// Press ']' via Update — should switch to next tab.
	bracketKey := keyMsgFromString("]")
	t.Logf("bracket key msg: Type=%d Runes=%v String=%q", bracketKey.Type, bracketKey.Runes, bracketKey.String())

	model, cmd = m.Update(bracketKey)
	m = model.(Model)

	// handleGlobalKeys should return non-nil cmd.
	if cmd == nil {
		t.Fatal("expected non-nil cmd from pressing ']'")
	}

	// Process the SwitchTabMsg.
	msg := cmd()
	t.Logf("cmd produced msg: %T %+v", msg, msg)
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.tabs.ActiveID() != 1 {
		t.Fatalf("expected active tab 1 after pressing ']', got %d", m.tabs.ActiveID())
	}

	// Press '[' — should switch back.
	model, cmd = m.Update(keyMsgFromString("["))
	m = model.(Model)
	if cmd == nil {
		t.Fatal("expected non-nil cmd from pressing '['")
	}
	model, _ = m.Update(cmd())
	m = model.(Model)
	if m.tabs.ActiveID() != 0 {
		t.Fatalf("expected active tab 0 after pressing '[', got %d", m.tabs.ActiveID())
	}
}

// ---------------------------------------------------------------------------
// TestHLColumnSelection: h/l keys move selectedCol when results focused
// ---------------------------------------------------------------------------

func TestHLColumnSelection(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New(cfg, nil, nil)

	// Give it a window size.
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(Model)

	// Simulate query result arriving — this focuses results.
	ts := m.activeTabState()
	ts.Results.SetResults(&adapter.QueryResult{
		Columns:  []adapter.ColumnMeta{{Name: "id"}, {Name: "name"}, {Name: "email"}},
		Rows:     [][]string{{"1", "Alice", "alice@example.com"}},
		RowCount: 1,
		IsSelect: true,
	})
	m.setFocus(PaneResults)

	if m.focusedPane != PaneResults {
		t.Fatalf("precondition: expected results focused, got pane %d", m.focusedPane)
	}
	if !ts.Results.Focused() {
		t.Fatal("precondition: results.focused should be true")
	}

	// selectedCol should start at 0.
	if ts.Results.SelectedCol() != 0 {
		t.Fatalf("precondition: selectedCol = %d, want 0", ts.Results.SelectedCol())
	}

	// Press 'l' — should move selectedCol to 1.
	model, _ = m.Update(keyMsgFromString("l"))
	m = model.(Model)
	ts = m.activeTabState()
	if ts.Results.SelectedCol() != 1 {
		t.Fatalf("after 'l': selectedCol = %d, want 1", ts.Results.SelectedCol())
	}

	// Press 'l' again — should move to 2.
	model, _ = m.Update(keyMsgFromString("l"))
	m = model.(Model)
	ts = m.activeTabState()
	if ts.Results.SelectedCol() != 2 {
		t.Fatalf("after second 'l': selectedCol = %d, want 2", ts.Results.SelectedCol())
	}

	// Press 'l' at last column — should stay at 2.
	model, _ = m.Update(keyMsgFromString("l"))
	m = model.(Model)
	ts = m.activeTabState()
	if ts.Results.SelectedCol() != 2 {
		t.Fatalf("after third 'l' at boundary: selectedCol = %d, want 2", ts.Results.SelectedCol())
	}

	// Press 'h' — should move back to 1.
	model, _ = m.Update(keyMsgFromString("h"))
	m = model.(Model)
	ts = m.activeTabState()
	if ts.Results.SelectedCol() != 1 {
		t.Fatalf("after 'h': selectedCol = %d, want 1", ts.Results.SelectedCol())
	}
}
