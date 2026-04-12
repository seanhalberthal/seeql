package connmgr

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seanhalberthal/seeql/internal/config"
	"github.com/seanhalberthal/seeql/internal/theme"
)

func init() {
	theme.Current = theme.Default()
}

func TestNew(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test-pg", DSN: "postgres://localhost/testdb"},
	}
	m := New(conns)

	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(m.connections))
	}
	if m.visible {
		t.Fatal("expected not visible initially")
	}
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect, got %d", m.state)
	}
	if m.editing != -1 {
		t.Fatalf("expected editing=-1, got %d", m.editing)
	}
}

func TestNew_Empty(t *testing.T) {
	m := New(nil)
	if len(m.connections) != 0 {
		t.Fatalf("expected 0 connections, got %d", len(m.connections))
	}
}

func TestShowAndHide(t *testing.T) {
	m := New(nil)

	if m.Visible() {
		t.Fatal("expected not visible initially")
	}

	m.Show()
	if !m.Visible() {
		t.Fatal("expected visible after Show()")
	}
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect after Show, got %d", m.state)
	}
	if m.connFocus != focusDSN {
		t.Fatalf("expected focusDSN after Show, got %d", m.connFocus)
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor=0 after Show, got %d", m.cursor)
	}

	m.Hide()
	if m.Visible() {
		t.Fatal("expected not visible after Hide()")
	}
}

func TestInit(t *testing.T) {
	m := New(nil)
	cmd := m.Init()
	if cmd != nil {
		t.Fatal("expected nil cmd from Init")
	}
}

func TestSetSize(t *testing.T) {
	m := New(nil)
	m.SetSize(120, 40)

	if m.width != 120 {
		t.Fatalf("expected width=120, got %d", m.width)
	}
	if m.height != 40 {
		t.Fatalf("expected height=40, got %d", m.height)
	}
}

func TestConnections(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
	}
	m := New(conns)

	got := m.Connections()
	if len(got) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(got))
	}
}

func TestSetConnections(t *testing.T) {
	m := New(nil)
	m.SetConnections([]config.SavedConnection{{Name: "new", DSN: "postgres://new"}})

	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(m.connections))
	}
	if m.connections[0].Name != "new" {
		t.Fatalf("expected name='new', got %q", m.connections[0].Name)
	}
}

func TestView_NotVisible(t *testing.T) {
	m := New(nil)
	view := m.View()
	if view != "" {
		t.Fatalf("expected empty view when not visible, got %q", view)
	}
}

func TestView_ConnectState(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test-db", DSN: "postgres://localhost/testdb"},
	}
	m := New(conns)
	m.Show()

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view in connect state")
	}
}

func TestView_FormState(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view in form state")
	}
}

func TestView_TestingState(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateTesting

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view in testing state")
	}
}

// --- StateConnect: DSN input focused ---

func TestConnect_EnterWithDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m.dsnInput.SetValue("postgres://localhost/test")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Visible() {
		t.Fatal("expected not visible after enter with DSN")
	}
	if cmd == nil {
		t.Fatal("expected cmd from enter")
	}
	msg := cmd()
	connMsg, ok := msg.(ConnectRequestMsg)
	if !ok {
		t.Fatalf("expected ConnectRequestMsg, got %T", msg)
	}
	if connMsg.AdapterName != "postgres" {
		t.Fatalf("expected adapter 'postgres', got %q", connMsg.AdapterName)
	}
	if connMsg.DSN != "postgres://localhost/test" {
		t.Fatalf("expected DSN, got %q", connMsg.DSN)
	}
}

func TestConnect_EnterEmptyDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.Visible() {
		t.Fatal("expected still visible when DSN is empty")
	}
	if cmd != nil {
		t.Fatal("expected nil cmd for empty DSN")
	}
}

func TestConnect_EnterUnknownDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m.dsnInput.SetValue("unknownformat://whatever")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.Visible() {
		t.Fatal("expected still visible for undetectable DSN")
	}
	if !m.isError {
		t.Fatal("expected error message for undetectable DSN")
	}
}

func TestConnect_EscapeCloses(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.Visible() {
		t.Fatal("expected not visible after escape from DSN focus")
	}
}

func TestConnect_TabToList(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
	}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.connFocus != focusList {
		t.Fatalf("expected focusList after tab, got %d", m.connFocus)
	}
}

func TestConnect_TabNoSavedConnections(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.connFocus != focusDSN {
		t.Fatalf("expected to stay on focusDSN with no saved connections, got %d", m.connFocus)
	}
}

func TestConnect_TabBackToDSN(t *testing.T) {
	conns := []config.SavedConnection{{Name: "a", DSN: "postgres://a"}}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.connFocus != focusList {
		t.Fatal("expected focusList after first tab")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.connFocus != focusDSN {
		t.Fatal("expected focusDSN after second tab")
	}
}

func TestConnect_CtrlSOpensForm(t *testing.T) {
	m := New(nil)
	m.Show()

	m.dsnInput.SetValue("postgres://localhost/db")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after Ctrl+S, got %d", m.state)
	}
	if m.editing != -1 {
		t.Fatalf("expected editing=-1 for new, got %d", m.editing)
	}
	if m.formDSN.Value() != "postgres://localhost/db" {
		t.Fatalf("expected formDSN pre-filled, got %q", m.formDSN.Value())
	}
}

func TestConnect_CtrlSEmptyDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateConnect {
		t.Fatal("expected to stay in StateConnect with empty DSN")
	}
	if !m.isError {
		t.Fatal("expected error message for empty DSN")
	}
}

func TestConnect_CtrlTEmptyDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if m.state != StateConnect {
		t.Fatal("expected to stay in StateConnect with empty DSN for test")
	}
	if !m.isError {
		t.Fatal("expected error for empty DSN")
	}
}

func TestConnect_CtrlTWithDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m.dsnInput.SetValue("postgres://localhost/db")
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if m.state != StateTesting {
		t.Fatalf("expected StateTesting, got %d", m.state)
	}
	if m.prevState != StateConnect {
		t.Fatalf("expected prevState=StateConnect, got %d", m.prevState)
	}
	if cmd == nil {
		t.Fatal("expected cmd for test")
	}
}

func TestConnect_DSNParsedLive(t *testing.T) {
	m := New(nil)
	m.Show()

	m.dsnInput.SetValue("postgres://user@localhost:5432/mydb?sslmode=disable")
	m.parsed = ParseDSN(m.dsnInput.Value())

	if m.parsed.Adapter != "postgres" {
		t.Fatalf("expected parsed adapter=postgres, got %q", m.parsed.Adapter)
	}
	if m.parsed.Params["sslmode"] != "disable" {
		t.Fatalf("expected sslmode=disable in parsed params, got %q", m.parsed.Params["sslmode"])
	}
}

// --- StateConnect: list focused ---

func TestConnectList_Navigation(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
		{Name: "c", DSN: "postgres://c"},
	}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after j, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("expected cursor=2 after j, got %d", m.cursor)
	}

	// Should not go past the last item
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("expected cursor=2 at boundary, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after k, got %d", m.cursor)
	}
}

func TestConnectList_Enter(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test", DSN: "postgres://localhost/test"},
	}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Visible() {
		t.Fatal("expected not visible after enter on saved connection")
	}
	if cmd == nil {
		t.Fatal("expected cmd from enter")
	}
	msg := cmd()
	connMsg, ok := msg.(ConnectRequestMsg)
	if !ok {
		t.Fatalf("expected ConnectRequestMsg, got %T", msg)
	}
	if connMsg.DSN != "postgres://localhost/test" {
		t.Fatalf("expected DSN, got %q", connMsg.DSN)
	}
}

func TestConnectList_NewConnection(t *testing.T) {
	conns := []config.SavedConnection{{Name: "a", DSN: "postgres://a"}}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after 'n', got %d", m.state)
	}
	if m.editing != -1 {
		t.Fatalf("expected editing=-1 for new, got %d", m.editing)
	}
}

func TestConnectList_EditConnection(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test", DSN: "postgres://localhost/test"},
	}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after 'e', got %d", m.state)
	}
	if m.editing != 0 {
		t.Fatalf("expected editing=0, got %d", m.editing)
	}
	if m.nameInput.Value() != "test" {
		t.Fatalf("expected name='test', got %q", m.nameInput.Value())
	}
	if m.formDSN.Value() != "postgres://localhost/test" {
		t.Fatalf("expected formDSN loaded, got %q", m.formDSN.Value())
	}
}

func TestConnectList_DeleteConnection(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
	}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection after delete, got %d", len(m.connections))
	}
	if m.connections[0].Name != "b" {
		t.Fatalf("expected remaining connection 'b', got %q", m.connections[0].Name)
	}
}

func TestConnectList_DeleteLastConnection(t *testing.T) {
	conns := []config.SavedConnection{{Name: "only", DSN: "postgres://only"}}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(m.connections) != 0 {
		t.Fatalf("expected 0 connections after delete, got %d", len(m.connections))
	}
}

func TestConnectList_EscapeBackToDSN(t *testing.T) {
	conns := []config.SavedConnection{{Name: "a", DSN: "postgres://a"}}
	m := New(conns)
	m.Show()
	m.connFocus = focusList
	m.dsnInput.Blur()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.connFocus != focusDSN {
		t.Fatalf("expected focusDSN after escape from list, got %d", m.connFocus)
	}
	if !m.Visible() {
		t.Fatal("expected still visible after escape from list (should return to DSN)")
	}
}

// --- StateForm ---

func TestForm_Escape(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect after escape, got %d", m.state)
	}
}

func TestForm_TabNavigation(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formFocus = 0

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.formFocus != 1 {
		t.Fatalf("expected formFocus=1 after tab, got %d", m.formFocus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.formFocus != 0 {
		t.Fatalf("expected formFocus=0 after second tab (wrap), got %d", m.formFocus)
	}
}

func TestForm_ShiftTabNavigation(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formFocus = 1

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.formFocus != 0 {
		t.Fatalf("expected formFocus=0 after shift+tab, got %d", m.formFocus)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.formFocus != 1 {
		t.Fatalf("expected formFocus=1 after shift+tab wrap, got %d", m.formFocus)
	}
}

func TestForm_SaveNew(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.editing = -1
	m.nameInput.SetValue("new-conn")
	m.formDSN.SetValue("postgres://localhost/newdb")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect after save, got %d", m.state)
	}
	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection after save, got %d", len(m.connections))
	}
	if m.connections[0].Name != "new-conn" {
		t.Fatalf("expected name 'new-conn', got %q", m.connections[0].Name)
	}
	if m.connections[0].DSN != "postgres://localhost/newdb" {
		t.Fatalf("expected DSN set, got %q", m.connections[0].DSN)
	}
}

func TestForm_SaveRequiresDSN(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.editing = -1
	m.nameInput.SetValue("no-dsn")
	// Leave DSN empty

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateForm {
		t.Fatalf("expected to stay in StateForm when DSN is empty, got %d", m.state)
	}
	if !m.isError {
		t.Fatal("expected error message when DSN is empty")
	}
}

func TestForm_SaveEdit(t *testing.T) {
	conns := []config.SavedConnection{{Name: "old", DSN: "postgres://old"}}
	m := New(conns)
	m.Show()
	m.state = StateForm
	m.editing = 0
	m.nameInput.SetValue("updated")
	m.formDSN.SetValue("postgres://old")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection after edit save, got %d", len(m.connections))
	}
	if m.connections[0].Name != "updated" {
		t.Fatalf("expected name 'updated', got %q", m.connections[0].Name)
	}
}

func TestForm_CtrlTEmptyDSN(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if m.state != StateForm {
		t.Fatal("expected to stay in StateForm with empty DSN")
	}
	if !m.isError {
		t.Fatal("expected error for empty DSN")
	}
}

func TestForm_CtrlTWithDSN(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formDSN.SetValue("postgres://localhost/db")

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if m.state != StateTesting {
		t.Fatalf("expected StateTesting, got %d", m.state)
	}
	if m.prevState != StateForm {
		t.Fatalf("expected prevState=StateForm, got %d", m.prevState)
	}
	if cmd == nil {
		t.Fatal("expected cmd for test")
	}
}

// --- StateTesting ---

func TestUpdateTesting_Success(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateTesting
	m.prevState = StateConnect

	m, _ = m.Update(testResultMsg{err: nil})
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect after success, got %d", m.state)
	}
	if m.isError {
		t.Fatal("expected isError=false for success")
	}
	if m.message != "Connection successful!" {
		t.Fatalf("unexpected message: %q", m.message)
	}
}

func TestUpdateTesting_Error(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateTesting
	m.prevState = StateForm

	m, _ = m.Update(testResultMsg{err: fmt.Errorf("conn refused")})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after error, got %d", m.state)
	}
	if !m.isError {
		t.Fatal("expected isError=true for failure")
	}
}

func TestUpdateTesting_Escape(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateTesting
	m.prevState = StateConnect

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.state != StateConnect {
		t.Fatalf("expected StateConnect after escape in testing, got %d", m.state)
	}
}

func TestUpdate_NotVisible(t *testing.T) {
	m := New(nil)
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Fatal("expected nil cmd when not visible")
	}
}

func TestDialogWidth(t *testing.T) {
	m := New(nil)
	if w := m.dialogWidth(); w != 64 {
		t.Fatalf("expected dialogWidth=64, got %d", w)
	}

	m.width = 50
	if w := m.dialogWidth(); w != 46 {
		t.Fatalf("expected dialogWidth=46 for width=50, got %d", w)
	}
}

func TestTruncateDSN(t *testing.T) {
	short := "postgres://***@localhost/db"
	got := truncateDSN("postgres://user:pass@localhost/db", 50)
	if got != short {
		t.Fatalf("expected sanitised DSN %q, got %q", short, got)
	}

	long := "postgres://user:pass@really-long-host.example.com:5432/some_database_name?sslmode=disable"
	got = truncateDSN(long, 40)
	if len(got) > 40 {
		t.Fatalf("expected truncated to 40 chars, got %d: %q", len(got), got)
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"postgres://user:pass@host/db",
			"postgres://***@host/db",
		},
		{
			"postgresql://admin:secret@host/db",
			"postgresql://***@host/db",
		},
		{
			"mysql://root:topsecret@host/db",
			"mysql://***@host/db",
		},
		{
			"password=secret123 host=localhost",
			"password=*** host=localhost",
		},
		{
			"no credentials here",
			"no credentials here",
		},
	}
	for _, tt := range tests {
		got := sanitizeError(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeError(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
