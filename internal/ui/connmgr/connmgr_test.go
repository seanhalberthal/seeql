package connmgr

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sadopc/seeql/internal/config"
	"github.com/sadopc/seeql/internal/theme"
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
	if m.state != StateList {
		t.Fatalf("expected StateList, got %d", m.state)
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
	if m.state != StateList {
		t.Fatalf("expected StateList after Show, got %d", m.state)
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

func TestView_ListState(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test-db", DSN: "postgres://localhost/testdb"},
	}
	m := New(conns)
	m.Show()

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view in list state")
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

func TestUpdateList_Navigation_JK(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
	}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after j, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("expected cursor=2 after j, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after k, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Fatalf("expected cursor=0 at boundary, got %d", m.cursor)
	}
}

func TestUpdateList_Navigation_ArrowKeys(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
	}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after down, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Fatalf("expected cursor=0 after up, got %d", m.cursor)
	}
}

func TestUpdateList_NewConnection(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after 'n', got %d", m.state)
	}
	if m.editing != -1 {
		t.Fatalf("expected editing=-1 for new connection, got %d", m.editing)
	}
}

func TestUpdateList_EditConnection(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test", DSN: "postgres://localhost/test"},
	}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after 'e', got %d", m.state)
	}
	if m.editing != 0 {
		t.Fatalf("expected editing=0, got %d", m.editing)
	}
	if m.inputs[fieldName].Value() != "test" {
		t.Fatalf("expected name field = 'test', got %q", m.inputs[fieldName].Value())
	}
	if m.inputs[fieldDSN].Value() != "postgres://localhost/test" {
		t.Fatalf("expected DSN field loaded, got %q", m.inputs[fieldDSN].Value())
	}
}

func TestUpdateList_DeleteConnection(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "a", DSN: "postgres://a"},
		{Name: "b", DSN: "postgres://b"},
	}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection after delete, got %d", len(m.connections))
	}
	if m.connections[0].Name != "b" {
		t.Fatalf("expected remaining connection 'b', got %q", m.connections[0].Name)
	}
}

func TestUpdateList_DeleteLastConnection(t *testing.T) {
	conns := []config.SavedConnection{{Name: "only", DSN: "postgres://only"}}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(m.connections) != 0 {
		t.Fatalf("expected 0 connections after delete, got %d", len(m.connections))
	}
}

func TestUpdateList_Enter_Connect(t *testing.T) {
	conns := []config.SavedConnection{
		{Name: "test", DSN: "postgres://localhost/test"},
	}
	m := New(conns)
	m.Show()

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Visible() {
		t.Fatal("expected not visible after enter (connecting)")
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

func TestUpdateList_Escape(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.Visible() {
		t.Fatal("expected not visible after escape")
	}
}

func TestUpdateList_Q(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.Visible() {
		t.Fatal("expected not visible after 'q'")
	}
}

func TestUpdateForm_Escape(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.state != StateList {
		t.Fatalf("expected StateList after escape, got %d", m.state)
	}
}

func TestUpdateForm_TabNavigation(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formFocus = 0

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.formFocus != 1 {
		t.Fatalf("expected formFocus=1 after tab, got %d", m.formFocus)
	}
}

func TestUpdateForm_ShiftTabNavigation(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formFocus = 1

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.formFocus != 0 {
		t.Fatalf("expected formFocus=0 after shift+tab, got %d", m.formFocus)
	}
}

func TestUpdateForm_ShiftTabWrap(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateForm
	m.formFocus = 0

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.formFocus != fieldCount-1 {
		t.Fatalf("expected formFocus=%d after shift+tab wrap, got %d", fieldCount-1, m.formFocus)
	}
}

func TestUpdateForm_SaveNew(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m.inputs[fieldName].SetValue("new-conn")
	m.inputs[fieldDSN].SetValue("postgres://localhost/newdb")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateList {
		t.Fatalf("expected StateList after save, got %d", m.state)
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

func TestUpdateForm_SaveRequiresDSN(t *testing.T) {
	m := New(nil)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m.inputs[fieldName].SetValue("no-dsn")
	// Leave DSN empty

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.state != StateForm {
		t.Fatalf("expected to stay in StateForm when DSN is empty, got %d", m.state)
	}
	if !m.isError {
		t.Fatal("expected error message when DSN is empty")
	}
}

func TestUpdateForm_SaveEdit(t *testing.T) {
	conns := []config.SavedConnection{{Name: "old", DSN: "postgres://old"}}
	m := New(conns)
	m.Show()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m.inputs[fieldName].SetValue("updated")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if len(m.connections) != 1 {
		t.Fatalf("expected 1 connection after edit save, got %d", len(m.connections))
	}
	if m.connections[0].Name != "updated" {
		t.Fatalf("expected name 'updated', got %q", m.connections[0].Name)
	}
}

func TestUpdate_NotVisible(t *testing.T) {
	m := New(nil)
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Fatal("expected nil cmd when not visible")
	}
}

func TestFormToConnection(t *testing.T) {
	m := New(nil)
	m.inputs[fieldName].SetValue("test")
	m.inputs[fieldDSN].SetValue("postgres://admin:secret@localhost:5432/mydb")

	conn := m.formToConnection()
	if conn.Name != "test" {
		t.Fatalf("expected name 'test', got %q", conn.Name)
	}
	if conn.DSN != "postgres://admin:secret@localhost:5432/mydb" {
		t.Fatalf("expected DSN, got %q", conn.DSN)
	}
}

func TestDialogWidth(t *testing.T) {
	m := New(nil)
	if w := m.dialogWidth(); w != 60 {
		t.Fatalf("expected dialogWidth=60, got %d", w)
	}

	m.width = 50
	if w := m.dialogWidth(); w != 46 {
		t.Fatalf("expected dialogWidth=46 for width=50, got %d", w)
	}
}

func TestUpdateTesting_Success(t *testing.T) {
	m := New(nil)
	m.Show()
	m.state = StateTesting

	m, _ = m.Update(testResultMsg{err: nil})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after success, got %d", m.state)
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

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.state != StateForm {
		t.Fatalf("expected StateForm after escape in testing, got %d", m.state)
	}
}

func TestClearForm(t *testing.T) {
	m := New(nil)
	m.inputs[fieldName].SetValue("something")
	m.inputs[fieldDSN].SetValue("postgres://localhost")
	m.message = "old message"

	m.clearForm()

	if m.inputs[fieldName].Value() != "" {
		t.Fatalf("expected name cleared, got %q", m.inputs[fieldName].Value())
	}
	if m.inputs[fieldDSN].Value() != "" {
		t.Fatalf("expected DSN cleared, got %q", m.inputs[fieldDSN].Value())
	}
	if m.formFocus != 0 {
		t.Fatalf("expected formFocus=0 after clear, got %d", m.formFocus)
	}
	if m.message != "" {
		t.Fatalf("expected message cleared, got %q", m.message)
	}
}

func TestLoadIntoForm(t *testing.T) {
	m := New(nil)
	conn := config.SavedConnection{
		Name: "mydb",
		DSN:  "postgres://admin:secret@db.example.com:3306/production",
	}

	m.loadIntoForm(conn)

	if m.inputs[fieldName].Value() != "mydb" {
		t.Fatalf("expected name 'mydb', got %q", m.inputs[fieldName].Value())
	}
	if m.inputs[fieldDSN].Value() != "postgres://admin:secret@db.example.com:3306/production" {
		t.Fatalf("expected DSN loaded, got %q", m.inputs[fieldDSN].Value())
	}
}
