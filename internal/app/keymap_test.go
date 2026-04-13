package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// containsKey checks whether the binding's keys contain the given key string.
func containsKey(b key.Binding, target string) bool {
	for _, k := range b.Keys() {
		if k == target {
			return true
		}
	}
	return false
}

// requireNonEmpty fails the test if the binding has no keys.
func requireNonEmpty(t *testing.T, name string, b key.Binding) {
	t.Helper()
	if len(b.Keys()) == 0 {
		t.Errorf("%s binding has no keys", name)
	}
}

// ---------------------------------------------------------------------------
// StandardKeyMap
// ---------------------------------------------------------------------------

func TestStandardKeyMap(t *testing.T) {
	km := StandardKeyMap()

	t.Run("ExecuteQuery has keys", func(t *testing.T) {
		requireNonEmpty(t, "ExecuteQuery", km.ExecuteQuery)
		if !containsKey(km.ExecuteQuery, "f5") {
			t.Error("ExecuteQuery should contain f5")
		}
	})

	t.Run("Quit has ctrl+q", func(t *testing.T) {
		requireNonEmpty(t, "Quit", km.Quit)
		if !containsKey(km.Quit, "ctrl+q") {
			t.Errorf("Quit keys = %v, want to contain %q", km.Quit.Keys(), "ctrl+q")
		}
	})

	t.Run("FocusNext has tab", func(t *testing.T) {
		requireNonEmpty(t, "FocusNext", km.FocusNext)
		if !containsKey(km.FocusNext, "tab") {
			t.Errorf("FocusNext keys = %v, want to contain %q", km.FocusNext.Keys(), "tab")
		}
	})

	t.Run("FocusPrev has shift+tab", func(t *testing.T) {
		requireNonEmpty(t, "FocusPrev", km.FocusPrev)
		if !containsKey(km.FocusPrev, "shift+tab") {
			t.Errorf("FocusPrev keys = %v, want to contain %q", km.FocusPrev.Keys(), "shift+tab")
		}
	})

	t.Run("NewTab has ctrl+t", func(t *testing.T) {
		requireNonEmpty(t, "NewTab", km.NewTab)
		if !containsKey(km.NewTab, "ctrl+t") {
			t.Errorf("NewTab keys = %v, want to contain %q", km.NewTab.Keys(), "ctrl+t")
		}
	})

	t.Run("CloseTab has ctrl+w", func(t *testing.T) {
		requireNonEmpty(t, "CloseTab", km.CloseTab)
		if !containsKey(km.CloseTab, "ctrl+w") {
			t.Errorf("CloseTab keys = %v, want to contain %q", km.CloseTab.Keys(), "ctrl+w")
		}
	})

	t.Run("CancelQuery has ctrl+c", func(t *testing.T) {
		requireNonEmpty(t, "CancelQuery", km.CancelQuery)
		if !containsKey(km.CancelQuery, "ctrl+c") {
			t.Errorf("CancelQuery keys = %v, want to contain %q", km.CancelQuery.Keys(), "ctrl+c")
		}
	})

	t.Run("Help has f1", func(t *testing.T) {
		requireNonEmpty(t, "Help", km.Help)
		if !containsKey(km.Help, "f1") {
			t.Errorf("Help keys = %v, want to contain %q", km.Help.Keys(), "f1")
		}
	})

	t.Run("ToggleSidebar has ctrl+s", func(t *testing.T) {
		requireNonEmpty(t, "ToggleSidebar", km.ToggleSidebar)
		if !containsKey(km.ToggleSidebar, "ctrl+s") {
			t.Errorf("ToggleSidebar keys = %v, want to contain %q", km.ToggleSidebar.Keys(), "ctrl+s")
		}
	})

	t.Run("RefreshSchema has ctrl+r", func(t *testing.T) {
		requireNonEmpty(t, "RefreshSchema", km.RefreshSchema)
		if !containsKey(km.RefreshSchema, "ctrl+r") {
			t.Errorf("RefreshSchema keys = %v, want to contain %q", km.RefreshSchema.Keys(), "ctrl+r")
		}
	})

	t.Run("Export has ctrl+e", func(t *testing.T) {
		requireNonEmpty(t, "Export", km.Export)
		if !containsKey(km.Export, "ctrl+e") {
			t.Errorf("Export keys = %v, want to contain %q", km.Export.Keys(), "ctrl+e")
		}
	})

	t.Run("OpenConnMgr has ctrl+o", func(t *testing.T) {
		requireNonEmpty(t, "OpenConnMgr", km.OpenConnMgr)
		if !containsKey(km.OpenConnMgr, "ctrl+o") {
			t.Errorf("OpenConnMgr keys = %v, want to contain %q", km.OpenConnMgr.Keys(), "ctrl+o")
		}
	})

}

func TestStandardKeyMap_AllNavigationBindingsHaveKeys(t *testing.T) {
	km := StandardKeyMap()
	bindings := map[string]key.Binding{
		"FocusNext":    km.FocusNext,
		"FocusPrev":    km.FocusPrev,
		"FocusSidebar": km.FocusSidebar,
		"FocusEditor":  km.FocusEditor,
		"FocusResults": km.FocusResults,
	}
	for name, b := range bindings {
		requireNonEmpty(t, name, b)
	}
}

func TestStandardKeyMap_AllTabBindingsHaveKeys(t *testing.T) {
	km := StandardKeyMap()
	bindings := map[string]key.Binding{
		"NewTab":   km.NewTab,
		"CloseTab": km.CloseTab,
		"NextTab":  km.NextTab,
		"PrevTab":  km.PrevTab,
	}
	for name, b := range bindings {
		requireNonEmpty(t, name, b)
	}
}

func TestStandardKeyMap_AllResizeBindingsHaveKeys(t *testing.T) {
	km := StandardKeyMap()
	bindings := map[string]key.Binding{
		"ResizeLeft":  km.ResizeLeft,
		"ResizeRight": km.ResizeRight,
		"ResizeUp":    km.ResizeUp,
		"ResizeDown":  km.ResizeDown,
	}
	for name, b := range bindings {
		requireNonEmpty(t, name, b)
	}
}

// ---------------------------------------------------------------------------
// ShortHelp
// ---------------------------------------------------------------------------

func TestShortHelp(t *testing.T) {
	km := StandardKeyMap()
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Fatal("ShortHelp() returned empty slice")
	}
	// ShortHelp should return exactly 5 bindings: ExecuteQuery, FocusNext, NewTab, Quit, Help
	if len(short) != 5 {
		t.Errorf("ShortHelp() length = %d, want 5", len(short))
	}
	// Verify each binding has keys
	for i, b := range short {
		if len(b.Keys()) == 0 {
			t.Errorf("ShortHelp()[%d] has no keys", i)
		}
	}
}

// ---------------------------------------------------------------------------
// FullHelp
// ---------------------------------------------------------------------------

func TestFullHelp(t *testing.T) {
	km := StandardKeyMap()
	full := km.FullHelp()
	if len(full) == 0 {
		t.Fatal("FullHelp() returned empty slice")
	}
	// FullHelp returns 6 groups of bindings
	if len(full) != 6 {
		t.Errorf("FullHelp() groups = %d, want 6", len(full))
	}
	// Each group should be non-empty
	for i, group := range full {
		if len(group) == 0 {
			t.Errorf("FullHelp()[%d] is empty", i)
		}
		// Each binding in the group should have keys
		for j, b := range group {
			if len(b.Keys()) == 0 {
				t.Errorf("FullHelp()[%d][%d] has no keys", i, j)
			}
		}
	}
}

func TestFullHelp_ContainsAllGroups(t *testing.T) {
	km := StandardKeyMap()
	full := km.FullHelp()

	// Group 0: Editor actions (ExecuteQuery, CancelQuery, Export)
	if len(full[0]) != 3 {
		t.Errorf("FullHelp group 0 (editor) length = %d, want 3", len(full[0]))
	}
	// Group 1: Navigation (FocusNext, FocusPrev, FocusSidebar, FocusEditor, FocusResults)
	if len(full[1]) != 5 {
		t.Errorf("FullHelp group 1 (navigation) length = %d, want 5", len(full[1]))
	}
	// Group 2: Tabs (NewTab, CloseTab, NextTab, PrevTab)
	if len(full[2]) != 4 {
		t.Errorf("FullHelp group 2 (tabs) length = %d, want 4", len(full[2]))
	}
	// Group 3: App (ToggleSidebar, RefreshSchema, OpenConnMgr, History)
	if len(full[3]) != 4 {
		t.Errorf("FullHelp group 3 (app) length = %d, want 4", len(full[3]))
	}
	// Group 4: Resize (ResizeLeft, ResizeRight, ResizeUp, ResizeDown)
	if len(full[4]) != 4 {
		t.Errorf("FullHelp group 4 (resize) length = %d, want 4", len(full[4]))
	}
	// Group 5: Quit + Help
	if len(full[5]) != 2 {
		t.Errorf("FullHelp group 5 (quit/help) length = %d, want 2", len(full[5]))
	}
}

// ---------------------------------------------------------------------------
// Specific key values
// ---------------------------------------------------------------------------

func TestStandardKeyMap_SpecificKeyValues(t *testing.T) {
	km := StandardKeyMap()

	tests := []struct {
		name    string
		binding key.Binding
		wantKey string
	}{
		{"FocusNext", km.FocusNext, "tab"},
		{"FocusPrev", km.FocusPrev, "shift+tab"},
		{"FocusSidebar", km.FocusSidebar, "alt+1"},
		{"FocusEditor", km.FocusEditor, "alt+2"},
		{"FocusResults", km.FocusResults, "alt+3"},
		{"NewTab", km.NewTab, "ctrl+t"},
		{"CloseTab", km.CloseTab, "ctrl+w"},
		{"Quit", km.Quit, "ctrl+q"},
		{"Help", km.Help, "f1"},
		{"ToggleSidebar", km.ToggleSidebar, "ctrl+s"},
		{"RefreshSchema", km.RefreshSchema, "ctrl+r"},
		{"OpenConnMgr", km.OpenConnMgr, "ctrl+o"},
		{"Export", km.Export, "ctrl+e"},
		{"CancelQuery", km.CancelQuery, "ctrl+c"},
		{"ResizeLeft", km.ResizeLeft, "ctrl+left"},
		{"ResizeRight", km.ResizeRight, "ctrl+right"},
		{"ResizeUp", km.ResizeUp, "ctrl+up"},
		{"ResizeDown", km.ResizeDown, "ctrl+down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !containsKey(tt.binding, tt.wantKey) {
				t.Errorf("%s keys = %v, want to contain %q", tt.name, tt.binding.Keys(), tt.wantKey)
			}
		})
	}
}

func TestStandardKeyMap_ExecuteQueryMultipleKeys(t *testing.T) {
	km := StandardKeyMap()
	keys := km.ExecuteQuery.Keys()
	if len(keys) < 2 {
		t.Errorf("ExecuteQuery should have at least 2 keys (f5 and ctrl+g), got %v", keys)
	}
	if !containsKey(km.ExecuteQuery, "f5") {
		t.Errorf("ExecuteQuery missing f5, keys = %v", keys)
	}
}

func TestStandardKeyMap_NextTabMultipleKeys(t *testing.T) {
	km := StandardKeyMap()
	keys := km.NextTab.Keys()
	if len(keys) < 2 {
		t.Errorf("NextTab should have at least 2 keys, got %v", keys)
	}
	if !containsKey(km.NextTab, "ctrl+pgdown") {
		t.Errorf("NextTab missing ctrl+pgdown, keys = %v", keys)
	}
	if !containsKey(km.NextTab, "]") {
		t.Errorf("NextTab missing ], keys = %v", keys)
	}
}

func TestStandardKeyMap_PrevTabMultipleKeys(t *testing.T) {
	km := StandardKeyMap()
	keys := km.PrevTab.Keys()
	if len(keys) < 2 {
		t.Errorf("PrevTab should have at least 2 keys, got %v", keys)
	}
	if !containsKey(km.PrevTab, "ctrl+pgup") {
		t.Errorf("PrevTab missing ctrl+pgup, keys = %v", keys)
	}
	if !containsKey(km.PrevTab, "[") {
		t.Errorf("PrevTab missing [, keys = %v", keys)
	}
}
