package connmgr

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seanhalberthal/seeql/internal/adapter"
	"github.com/seanhalberthal/seeql/internal/config"
	"github.com/seanhalberthal/seeql/internal/theme"
)

// State tracks the connection manager screen.
type State int

const (
	StateConnect State = iota // Combined view: DSN input + saved list
	StateForm                 // Save/edit named connection
	StateTesting              // Testing connection
)

// Focus within StateConnect — either the DSN input or the saved list.
type connectFocus int

const (
	focusDSN  connectFocus = iota
	focusList
)

// ConnectRequestMsg is sent when the user picks a connection.
type ConnectRequestMsg struct {
	AdapterName string
	DSN         string
}

// ConnectionsUpdatedMsg is sent when saved connections are modified.
type ConnectionsUpdatedMsg struct {
	Connections []config.SavedConnection
}

// Model is the connection manager modal.
type Model struct {
	state       State
	connections []config.SavedConnection
	cursor      int
	visible     bool
	width       int
	height      int

	// DSN input (primary)
	dsnInput  textinput.Model
	parsed    ParsedDSN
	connFocus connectFocus

	// Form fields (save/edit)
	nameInput textinput.Model
	formDSN   textinput.Model
	formFocus int // 0=name, 1=dsn
	editing   int // index of connection being edited, -1 for new

	// Feedback
	message string
	isError bool

	// Track previous state for testing return
	prevState State
}

// New creates a new connection manager.
func New(connections []config.SavedConnection) Model {
	dsn := textinput.New()
	dsn.Prompt = "DSN: "
	dsn.Placeholder = "postgres://user:pass@host:5432/db?sslmode=disable"
	dsn.Width = 60
	dsn.CharLimit = 512

	name := textinput.New()
	name.Prompt = "Name: "
	name.Placeholder = "my-database (optional)"
	name.Width = 50

	formDSN := textinput.New()
	formDSN.Prompt = "DSN: "
	formDSN.Placeholder = "postgres://user:pass@host:5432/db"
	formDSN.Width = 50
	formDSN.CharLimit = 512

	return Model{
		connections: connections,
		editing:     -1,
		dsnInput:    dsn,
		nameInput:   name,
		formDSN:     formDSN,
		connFocus:   focusDSN,
	}
}

// Init returns no initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles connection manager messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch m.state {
	case StateConnect:
		return m.updateConnect(msg)
	case StateForm:
		return m.updateForm(msg)
	case StateTesting:
		return m.updateTesting(msg)
	}
	return m, nil
}

// --- StateConnect: combined DSN input + saved connections list ---

func (m Model) updateConnect(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.connFocus == focusList {
				m.connFocus = focusDSN
				m.dsnInput.Focus()
				return m, textinput.Blink
			}
			m.visible = false
			return m, nil

		case "tab":
			if m.connFocus == focusDSN {
				if len(m.connections) == 0 {
					return m, nil
				}
				m.connFocus = focusList
				m.dsnInput.Blur()
			} else {
				m.connFocus = focusDSN
				m.dsnInput.Focus()
				return m, textinput.Blink
			}
			return m, nil

		case "enter":
			if m.connFocus == focusDSN {
				dsn := strings.TrimSpace(m.dsnInput.Value())
				if dsn == "" {
					return m, nil
				}
				adapterName := adapter.DetectAdapter(dsn)
				if adapterName == "" {
					m.message = "Could not detect database type from DSN"
					m.isError = true
					return m, nil
				}
				m.visible = false
				return m, func() tea.Msg {
					return ConnectRequestMsg{AdapterName: adapterName, DSN: dsn}
				}
			}
			if m.cursor < len(m.connections) {
				conn := m.connections[m.cursor]
				adapterName := adapter.DetectAdapter(conn.DSN)
				m.visible = false
				return m, func() tea.Msg {
					return ConnectRequestMsg{AdapterName: adapterName, DSN: conn.DSN}
				}
			}
			return m, nil

		case "ctrl+s":
			dsn := strings.TrimSpace(m.dsnInput.Value())
			if dsn == "" {
				m.message = "DSN is required"
				m.isError = true
				return m, nil
			}
			m.state = StateForm
			m.editing = -1
			m.nameInput.SetValue("")
			m.formDSN.SetValue(dsn)
			m.formFocus = 0
			m.nameInput.Focus()
			m.formDSN.Blur()
			m.message = ""
			return m, textinput.Blink

		case "ctrl+t":
			dsn := strings.TrimSpace(m.dsnInput.Value())
			if dsn == "" {
				m.message = "DSN is required"
				m.isError = true
				return m, nil
			}
			m.prevState = StateConnect
			m.state = StateTesting
			return m, m.testConnection(config.SavedConnection{DSN: dsn})
		}

		// List navigation when focused on list
		if m.connFocus == focusList {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.connections)-1 {
					m.cursor++
				}
			case "n":
				m.state = StateForm
				m.editing = -1
				m.nameInput.SetValue("")
				m.formDSN.SetValue("")
				m.formFocus = 0
				m.nameInput.Focus()
				m.formDSN.Blur()
				m.message = ""
				return m, textinput.Blink
			case "e":
				if m.cursor < len(m.connections) {
					m.state = StateForm
					m.editing = m.cursor
					m.nameInput.SetValue(m.connections[m.cursor].Name)
					m.formDSN.SetValue(m.connections[m.cursor].DSN)
					m.formFocus = 0
					m.nameInput.Focus()
					m.formDSN.Blur()
					m.message = ""
					return m, textinput.Blink
				}
			case "d":
				if m.cursor < len(m.connections) {
					m.connections = append(m.connections[:m.cursor], m.connections[m.cursor+1:]...)
					if m.cursor >= len(m.connections) && m.cursor > 0 {
						m.cursor--
					}
					conns := make([]config.SavedConnection, len(m.connections))
					copy(conns, m.connections)
					return m, func() tea.Msg { return ConnectionsUpdatedMsg{Connections: conns} }
				}
			}
			return m, nil
		}

		// DSN input is focused — update it and reparse
		var cmd tea.Cmd
		m.dsnInput, cmd = m.dsnInput.Update(msg)
		m.parsed = ParseDSN(m.dsnInput.Value())
		m.message = ""
		return m, cmd

	default:
		if m.connFocus == focusDSN {
			var cmd tea.Cmd
			m.dsnInput, cmd = m.dsnInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// --- StateForm: save/edit named connection ---

func (m Model) updateForm(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = StateConnect
			m.dsnInput.Focus()
			return m, textinput.Blink
		case "tab", "down":
			if m.formFocus == 0 {
				m.nameInput.Blur()
				m.formDSN.Focus()
				m.formFocus = 1
			} else {
				m.formDSN.Blur()
				m.nameInput.Focus()
				m.formFocus = 0
			}
			return m, textinput.Blink
		case "shift+tab", "up":
			if m.formFocus == 1 {
				m.formDSN.Blur()
				m.nameInput.Focus()
				m.formFocus = 0
			} else {
				m.nameInput.Blur()
				m.formDSN.Focus()
				m.formFocus = 1
			}
			return m, textinput.Blink
		case "ctrl+s":
			dsn := strings.TrimSpace(m.formDSN.Value())
			if dsn == "" {
				m.message = "DSN is required"
				m.isError = true
				return m, nil
			}
			conn := config.SavedConnection{
				Name: strings.TrimSpace(m.nameInput.Value()),
				DSN:  dsn,
			}
			if m.editing >= 0 && m.editing < len(m.connections) {
				m.connections[m.editing] = conn
			} else {
				m.connections = append(m.connections, conn)
			}
			m.state = StateConnect
			m.dsnInput.Focus()
			conns := make([]config.SavedConnection, len(m.connections))
			copy(conns, m.connections)
			return m, tea.Batch(
				textinput.Blink,
				func() tea.Msg { return ConnectionsUpdatedMsg{Connections: conns} },
			)
		case "ctrl+t":
			dsn := strings.TrimSpace(m.formDSN.Value())
			if dsn == "" {
				m.message = "DSN is required"
				m.isError = true
				return m, nil
			}
			m.prevState = StateForm
			m.state = StateTesting
			return m, m.testConnection(config.SavedConnection{DSN: dsn})
		}
	}

	var cmd tea.Cmd
	if m.formFocus == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.formDSN, cmd = m.formDSN.Update(msg)
	}
	return m, cmd
}

// --- StateTesting ---

func (m Model) updateTesting(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case testResultMsg:
		if msg.err != nil {
			m.message = "Connection failed: " + sanitizeError(msg.err.Error())
			m.isError = true
		} else {
			m.message = "Connection successful!"
			m.isError = false
		}
		m.state = m.prevState
		if m.state == StateConnect {
			m.dsnInput.Focus()
			return m, textinput.Blink
		}
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.state = m.prevState
		}
	}
	return m, nil
}

type testResultMsg struct{ err error }

func (m Model) testConnection(conn config.SavedConnection) tea.Cmd {
	return func() tea.Msg {
		adapterName := adapter.DetectAdapter(conn.DSN)
		a, ok := adapter.Registry[adapterName]
		if !ok {
			return testResultMsg{err: fmt.Errorf("unknown adapter for DSN (detected: %q)", adapterName)}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		c, err := a.Connect(ctx, conn.DSN)
		if err != nil {
			return testResultMsg{err: err}
		}
		err = c.Ping(ctx)
		_ = c.Close()
		return testResultMsg{err: err}
	}
}

// --- Views ---

// View renders the connection manager.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	th := theme.Current

	switch m.state {
	case StateConnect:
		return m.viewConnect(th)
	case StateForm:
		return m.viewForm(th)
	case StateTesting:
		return th.DialogBorder.Render("\n  Testing connection...\n")
	}
	return ""
}

func (m Model) viewConnect(th *theme.Theme) string {
	title := th.DialogTitle.Render("  Connections  ")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	// DSN input
	lines = append(lines, "  "+m.dsnInput.View())

	// Parsed DSN summary
	if summary := m.parsed.Summary(); summary != "" {
		check := th.SuccessText.Render("\u2713")
		lines = append(lines, "  "+check+" "+th.MutedText.Render(summary))
		if params := m.parsed.ParamString(); params != "" {
			lines = append(lines, "    "+th.MutedText.Render(params))
		}
	}

	// Feedback message
	if m.message != "" {
		if m.isError {
			lines = append(lines, "  "+th.ErrorText.Render(m.message))
		} else {
			lines = append(lines, "  "+th.SuccessText.Render(m.message))
		}
	}

	// Divider + saved connections
	if len(m.connections) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+th.MutedText.Render("Saved"))

		for i, conn := range m.connections {
			label := conn.DSN
			if conn.Name != "" {
				label = conn.Name + "  " + th.MutedText.Render("("+truncateDSN(conn.DSN, 40)+")")
			}
			line := "  " + label
			if m.connFocus == focusList && i == m.cursor {
				lines = append(lines, th.SidebarSelected.Render(line))
			} else {
				lines = append(lines, "  "+line)
			}
		}
	}

	lines = append(lines, "")
	help := "enter:connect  ctrl+s:save  ctrl+t:test  esc:close"
	if m.connFocus == focusList {
		help = "enter:connect  n:new  e:edit  d:delete  tab:dsn  esc:back"
	}
	lines = append(lines, th.MutedText.Render("  "+help))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return th.DialogBorder.Width(m.dialogWidth()).Render(content)
}

func (m Model) viewForm(th *theme.Theme) string {
	title := "  New Connection  "
	if m.editing >= 0 {
		title = "  Edit Connection  "
	}

	var lines []string
	lines = append(lines, th.DialogTitle.Render(title))
	lines = append(lines, "")
	lines = append(lines, "  "+m.nameInput.View())
	lines = append(lines, "  "+m.formDSN.View())

	if m.message != "" {
		if m.isError {
			lines = append(lines, "", "  "+th.ErrorText.Render(m.message))
		} else {
			lines = append(lines, "", "  "+th.SuccessText.Render(m.message))
		}
	}

	lines = append(lines, "")
	lines = append(lines, th.MutedText.Render("  ctrl+s:save  ctrl+t:test  esc:back"))

	content := strings.Join(lines, "\n")
	return th.DialogBorder.Width(m.dialogWidth()).Render(content)
}

func (m Model) dialogWidth() int {
	w := 64
	if m.width > 0 && w > m.width-4 {
		w = m.width - 4
	}
	return w
}

// truncateDSN shortens a DSN for display, masking credentials.
func truncateDSN(dsn string, maxLen int) string {
	masked := sanitizeError(dsn)
	if len(masked) <= maxLen {
		return masked
	}
	return masked[:maxLen-1] + "\u2026"
}

// Show makes the connection manager visible with DSN input focused.
func (m *Model) Show() {
	m.visible = true
	m.state = StateConnect
	m.connFocus = focusDSN
	m.cursor = 0
	m.message = ""
	m.dsnInput.Focus()
}

// Hide hides the connection manager.
func (m *Model) Hide() {
	m.visible = false
	m.dsnInput.Blur()
}

// Visible returns whether the connection manager is shown.
func (m Model) Visible() bool { return m.visible }

// SetSize sets the available space.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Connections returns the current saved connections.
func (m Model) Connections() []config.SavedConnection {
	return m.connections
}

// SetConnections updates the saved connections list.
func (m *Model) SetConnections(conns []config.SavedConnection) {
	m.connections = conns
}

// sanitizeError strips credentials from error messages that may contain DSN URLs.
func sanitizeError(msg string) string {
	for _, prefix := range []string{"postgres://", "postgresql://", "mysql://"} {
		offset := 0
		for {
			idx := strings.Index(msg[offset:], prefix)
			if idx < 0 {
				break
			}
			idx += offset // absolute position
			rest := msg[idx+len(prefix):]
			atIdx := strings.Index(rest, "@")
			if atIdx < 0 {
				break
			}
			msg = msg[:idx+len(prefix)] + "***" + msg[idx+len(prefix)+atIdx:]
			offset = idx + len(prefix) + 4 // skip past "***@"
		}
	}
	msg = reMySQLCreds.ReplaceAllString(msg, "${1}***@tcp(")
	msg = rePGPassword.ReplaceAllString(msg, "password=***")
	return msg
}

var (
	reMySQLCreds = regexp.MustCompile(`(\b\w+:)[^@]+@tcp\(`)
	rePGPassword = regexp.MustCompile(`password=[^\s]+`)
)
