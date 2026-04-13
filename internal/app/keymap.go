package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all application keybindings.
type KeyMap struct {
	// Navigation
	FocusNext    key.Binding
	FocusPrev    key.Binding
	FocusSidebar key.Binding
	FocusEditor  key.Binding
	FocusResults key.Binding

	// Tabs
	NewTab   key.Binding
	CloseTab key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding

	// Editor
	ExecuteQuery key.Binding
	CancelQuery  key.Binding

	// App
	Quit          key.Binding
	Help          key.Binding
	ToggleSidebar key.Binding
	RefreshSchema key.Binding
	OpenConnMgr   key.Binding
	History       key.Binding
	Export        key.Binding

	// Pane resizing
	ResizeLeft  key.Binding
	ResizeRight key.Binding
	ResizeUp    key.Binding
	ResizeDown  key.Binding
}

// StandardKeyMap returns keybindings for standard mode.
func StandardKeyMap() KeyMap {
	return KeyMap{
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab", "ctrl+j"),
			key.WithHelp("shift+tab/ctrl+j", "prev pane"),
		),
		FocusSidebar: key.NewBinding(
			key.WithKeys("alt+1"),
			key.WithHelp("alt+1", "sidebar"),
		),
		FocusEditor: key.NewBinding(
			key.WithKeys("alt+2"),
			key.WithHelp("alt+2", "editor"),
		),
		FocusResults: key.NewBinding(
			key.WithKeys("alt+3"),
			key.WithHelp("alt+3", "results"),
		),
		NewTab: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "new tab"),
		),
		CloseTab: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("ctrl+w", "close tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("ctrl+pgdown", "]"),
			key.WithHelp("]", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("ctrl+pgup", "["),
			key.WithHelp("[", "prev tab"),
		),
		ExecuteQuery: key.NewBinding(
			key.WithKeys("f5", "ctrl+g"),
			key.WithHelp("f5", "run query"),
		),
		CancelQuery: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel query"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("f1"),
			key.WithHelp("f1", "help"),
		),
		ToggleSidebar: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "toggle sidebar"),
		),
		RefreshSchema: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh schema"),
		),
		OpenConnMgr: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "connections"),
		),
		History: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "history"),
		),
		Export: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "export"),
		),
		ResizeLeft: key.NewBinding(
			key.WithKeys("ctrl+left"),
			key.WithHelp("ctrl+←", "shrink left"),
		),
		ResizeRight: key.NewBinding(
			key.WithKeys("ctrl+right"),
			key.WithHelp("ctrl+→", "grow right"),
		),
		ResizeUp: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("ctrl+↑", "shrink up"),
		),
		ResizeDown: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("ctrl+↓", "grow down"),
		),
	}
}

// ShortHelp returns a subset of keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ExecuteQuery, k.FocusNext, k.NewTab, k.Quit, k.Help,
	}
}

// FullHelp returns all keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ExecuteQuery, k.CancelQuery, k.Export},
		{k.FocusNext, k.FocusPrev, k.FocusSidebar, k.FocusEditor, k.FocusResults},
		{k.NewTab, k.CloseTab, k.NextTab, k.PrevTab},
		{k.ToggleSidebar, k.RefreshSchema, k.OpenConnMgr, k.History},
		{k.ResizeLeft, k.ResizeRight, k.ResizeUp, k.ResizeDown},
		{k.Quit, k.Help},
	}
}
