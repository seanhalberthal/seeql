// Package theme provides a centralized styling system for the seeql
// terminal UI. A single adaptive theme inherits the user's terminal
// colours so it looks right in any colour scheme.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme holds lipgloss.Style values for every UI element in the application.
type Theme struct {
	Name string

	// App-level
	AppBackground lipgloss.Style

	// Sidebar / Schema browser
	SidebarBorder     lipgloss.Style
	SidebarTitle      lipgloss.Style
	SidebarDatabase   lipgloss.Style
	SidebarSchema     lipgloss.Style
	SidebarTable      lipgloss.Style
	SidebarView       lipgloss.Style
	SidebarColumn     lipgloss.Style
	SidebarColumnType lipgloss.Style
	SidebarSelected   lipgloss.Style

	// Editor
	EditorBorder     lipgloss.Style
	EditorLineNumber lipgloss.Style
	EditorCursor     lipgloss.Style

	// SQL Syntax highlighting (ANSI 16-colour palette)
	SQLKeyword    lipgloss.Style
	SQLString     lipgloss.Style
	SQLNumber     lipgloss.Style
	SQLComment    lipgloss.Style
	SQLOperator   lipgloss.Style
	SQLFunction   lipgloss.Style
	SQLType       lipgloss.Style
	SQLIdentifier lipgloss.Style

	// Results table
	ResultsBorder      lipgloss.Style
	ResultsHeader      lipgloss.Style
	ResultsCell        lipgloss.Style
	ResultsCellAlt     lipgloss.Style
	ResultsSelectedRow lipgloss.Style
	ResultsNull        lipgloss.Style

	// Tab bar
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabBar      lipgloss.Style

	// Status bar
	StatusBar        lipgloss.Style
	StatusBarKey     lipgloss.Style
	StatusBarValue   lipgloss.Style
	StatusBarError   lipgloss.Style
	StatusBarSuccess lipgloss.Style

	// Autocomplete
	AutocompleteItem     lipgloss.Style
	AutocompleteSelected lipgloss.Style
	AutocompleteBorder   lipgloss.Style

	// Dialog/Modal
	DialogBorder       lipgloss.Style
	DialogTitle        lipgloss.Style
	DialogButton       lipgloss.Style
	DialogButtonActive lipgloss.Style

	// General
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
	ErrorText       lipgloss.Style
	SuccessText     lipgloss.Style
	WarningText     lipgloss.Style
	MutedText       lipgloss.Style
}

// ANSI colour shortcuts using the terminal's 16-colour palette.
// These inherit the user's theme colours (Dracula, Catppuccin, etc.).
var (
	ansiBlack        = lipgloss.Color("0")
	ansiRed          = lipgloss.Color("1")
	ansiGreen        = lipgloss.Color("2")
	ansiYellow       = lipgloss.Color("3")
	ansiBlue         = lipgloss.Color("4")
	ansiMagenta      = lipgloss.Color("5")
	ansiCyan         = lipgloss.Color("6")
	ansiWhite        = lipgloss.Color("7")
	ansiBrightBlack  = lipgloss.Color("8")
	ansiBrightRed    = lipgloss.Color("9")
	ansiBrightGreen  = lipgloss.Color("10")
	ansiBrightYellow = lipgloss.Color("11")
	ansiBrightBlue   = lipgloss.Color("12")
	ansiBrightCyan   = lipgloss.Color("14")
	ansiBrightWhite  = lipgloss.Color("15")
)

// Adaptive foreground colours — automatically pick the right shade for
// light vs dark terminal backgrounds.
var (
	fgDefault = lipgloss.AdaptiveColor{Light: "0", Dark: "7"}   // text
	fgMuted   = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}   // dim
	fgAccent  = lipgloss.AdaptiveColor{Light: "4", Dark: "12"}  // blue accent
	fgBorder  = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}   // border lines
)

// newAdaptiveTheme builds the single adaptive theme. It uses NoColor{}
// for backgrounds (terminal transparency shows through) and the ANSI
// 16-colour palette for foregrounds.
func newAdaptiveTheme() *Theme {
	// Transparent background — lets the terminal's own background show.
	noBg := lipgloss.NewStyle()

	return &Theme{
		Name: "adaptive",

		// App-level — fully transparent
		AppBackground: noBg,

		// Sidebar
		SidebarBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgBorder),
		SidebarTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(fgAccent).
			PaddingLeft(1),
		SidebarDatabase: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiYellow),
		SidebarSchema: lipgloss.NewStyle().
			Foreground(ansiCyan),
		SidebarTable: lipgloss.NewStyle().
			Foreground(ansiGreen),
		SidebarView: lipgloss.NewStyle().
			Foreground(ansiMagenta),
		SidebarColumn: lipgloss.NewStyle().
			Foreground(fgDefault),
		SidebarColumnType: lipgloss.NewStyle().
			Foreground(fgMuted).
			Italic(true),
		SidebarSelected: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBrightWhite).
			Background(fgAccent),

		// Editor
		EditorBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgBorder),
		EditorLineNumber: lipgloss.NewStyle().
			Foreground(fgMuted),
		EditorCursor: lipgloss.NewStyle().
			Reverse(true),

		// SQL Syntax highlighting — ANSI palette
		SQLKeyword: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBlue),
		SQLString: lipgloss.NewStyle().
			Foreground(ansiGreen),
		SQLNumber: lipgloss.NewStyle().
			Foreground(ansiMagenta),
		SQLComment: lipgloss.NewStyle().
			Italic(true).
			Foreground(ansiBrightBlack),
		SQLOperator: lipgloss.NewStyle().
			Foreground(fgDefault),
		SQLFunction: lipgloss.NewStyle().
			Foreground(ansiYellow),
		SQLType: lipgloss.NewStyle().
			Foreground(ansiCyan),
		SQLIdentifier: lipgloss.NewStyle().
			Foreground(fgDefault),

		// Results table — transparent backgrounds, padding for spacing
		ResultsBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgBorder),
		ResultsHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(fgAccent).
			Padding(0, 1),
		ResultsCell: lipgloss.NewStyle().
			Foreground(fgDefault).
			Padding(0, 1),
		ResultsCellAlt: lipgloss.NewStyle().
			Foreground(fgDefault).
			Background(ansiBlack).
			Padding(0, 1),
		ResultsSelectedRow: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBrightWhite).
			Background(fgAccent).
			Padding(0, 1),
		ResultsNull: lipgloss.NewStyle().
			Italic(true).
			Foreground(fgMuted),

		// Tab bar
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(fgAccent).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(false).
			BorderForeground(fgAccent).
			PaddingLeft(1).
			PaddingRight(1),
		TabInactive: lipgloss.NewStyle().
			Foreground(fgMuted).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(fgBorder).
			PaddingLeft(1).
			PaddingRight(1),
		TabBar: noBg,

		// Status bar — subtle, not heavy
		StatusBar: lipgloss.NewStyle().
			Foreground(fgMuted),
		StatusBarKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(fgAccent).
			PaddingLeft(1).
			PaddingRight(1),
		StatusBarValue: lipgloss.NewStyle().
			Foreground(fgDefault).
			PaddingLeft(1).
			PaddingRight(1),
		StatusBarError: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBrightRed),
		StatusBarSuccess: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBrightGreen),

		// Autocomplete
		AutocompleteItem: lipgloss.NewStyle().
			Foreground(fgDefault).
			Background(ansiBlack).
			PaddingLeft(1).
			PaddingRight(1),
		AutocompleteSelected: lipgloss.NewStyle().
			Foreground(ansiBrightWhite).
			Background(fgAccent).
			PaddingLeft(1).
			PaddingRight(1),
		AutocompleteBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgAccent),

		// Dialog/Modal — muted background panel
		DialogBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgAccent).
			Padding(1, 2),
		DialogTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(fgAccent),
		DialogButton: lipgloss.NewStyle().
			Foreground(fgDefault).
			Background(ansiBrightBlack).
			PaddingLeft(2).
			PaddingRight(2),
		DialogButtonActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiBrightWhite).
			Background(fgAccent).
			PaddingLeft(2).
			PaddingRight(2),

		// General
		FocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgAccent),
		UnfocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(fgBorder),
		ErrorText: lipgloss.NewStyle().
			Bold(true).
			Foreground(ansiRed),
		SuccessText: lipgloss.NewStyle().
			Foreground(ansiGreen),
		WarningText: lipgloss.NewStyle().
			Foreground(ansiYellow),
		MutedText: lipgloss.NewStyle().
			Foreground(fgMuted),
	}
}

// Current is the active theme — a single adaptive theme that inherits
// the terminal's colour scheme.
var Current = newAdaptiveTheme()

// Default returns the adaptive theme (kept for API compatibility).
func Default() *Theme {
	return newAdaptiveTheme()
}

// Get returns the adaptive theme regardless of name. The name parameter
// is accepted for backwards compatibility but ignored.
func Get(name string) *Theme {
	return Current
}
