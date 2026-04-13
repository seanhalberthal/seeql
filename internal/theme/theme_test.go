package theme

import (
	"testing"
)

func TestDefault(t *testing.T) {
	d := Default()
	if d == nil {
		t.Fatal("Default() returned nil")
	}
	if d.Name != "adaptive" {
		t.Errorf("Default().Name = %q, want %q", d.Name, "adaptive")
	}
}

func TestGet_ReturnsAdaptive(t *testing.T) {
	for _, name := range []string{"default", "light", "monokai", "nonexistent", ""} {
		t.Run(name, func(t *testing.T) {
			th := Get(name)
			if th == nil {
				t.Fatalf("Get(%q) returned nil", name)
			}
			if th.Name != "adaptive" {
				t.Errorf("Get(%q).Name = %q, want %q", name, th.Name, "adaptive")
			}
		})
	}
}

func TestCurrent_InitialValue(t *testing.T) {
	if Current == nil {
		t.Fatal("Current is nil at init")
	}
	if Current.Name != "adaptive" {
		t.Errorf("Current.Name = %q, want %q", Current.Name, "adaptive")
	}
}

func TestAdaptiveTheme_SQLStyles(t *testing.T) {
	d := Default()
	tests := []struct {
		name  string
		style func() string
	}{
		{"SQLKeyword", func() string { return d.SQLKeyword.Render("SELECT") }},
		{"SQLString", func() string { return d.SQLString.Render("'hello'") }},
		{"SQLNumber", func() string { return d.SQLNumber.Render("42") }},
		{"SQLComment", func() string { return d.SQLComment.Render("-- note") }},
		{"SQLOperator", func() string { return d.SQLOperator.Render("=") }},
		{"SQLFunction", func() string { return d.SQLFunction.Render("COUNT") }},
		{"SQLType", func() string { return d.SQLType.Render("INTEGER") }},
		{"SQLIdentifier", func() string { return d.SQLIdentifier.Render("users") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.style()
			if out == "" {
				t.Errorf("%s rendered empty string", tt.name)
			}
		})
	}
}

func TestAdaptiveTheme_UIStyles_NotEmpty(t *testing.T) {
	th := Current
	pairs := []struct {
		label string
		out   string
	}{
		{"TabActive", th.TabActive.Render("tab")},
		{"TabInactive", th.TabInactive.Render("tab")},
		{"StatusBar", th.StatusBar.Render("status")},
		{"StatusBarError", th.StatusBarError.Render("err")},
		{"DialogBorder", th.DialogBorder.Render("dlg")},
		{"FocusedBorder", th.FocusedBorder.Render("focused")},
		{"UnfocusedBorder", th.UnfocusedBorder.Render("unfocused")},
		{"ErrorText", th.ErrorText.Render("error")},
		{"SuccessText", th.SuccessText.Render("ok")},
		{"MutedText", th.MutedText.Render("muted")},
		{"SidebarSelected", th.SidebarSelected.Render("sel")},
		{"AutocompleteItem", th.AutocompleteItem.Render("item")},
		{"AutocompleteSelected", th.AutocompleteSelected.Render("sel")},
		{"ResultsHeader", th.ResultsHeader.Render("hdr")},
	}
	for _, p := range pairs {
		if p.out == "" {
			t.Errorf("%s rendered empty", p.label)
		}
	}
}
