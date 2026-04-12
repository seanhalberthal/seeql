package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.KeyMode != "vim" {
		t.Errorf("KeyMode = %q, want %q", cfg.KeyMode, "vim")
	}
	if cfg.Editor.TabSize != 4 {
		t.Errorf("Editor.TabSize = %d, want %d", cfg.Editor.TabSize, 4)
	}
	if cfg.Editor.ShowLineNumbers != true {
		t.Errorf("Editor.ShowLineNumbers = %v, want %v", cfg.Editor.ShowLineNumbers, true)
	}
	if cfg.Results.PageSize != 1000 {
		t.Errorf("Results.PageSize = %d, want %d", cfg.Results.PageSize, 1000)
	}
	if cfg.Results.MaxColumnWidth != 50 {
		t.Errorf("Results.MaxColumnWidth = %d, want %d", cfg.Results.MaxColumnWidth, 50)
	}
	if len(cfg.Connections) != 0 {
		t.Errorf("Connections length = %d, want 0", len(cfg.Connections))
	}
}

func TestLoadValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := `{
  "keymode": "vim",
  "editor": {
    "tab_size": 2,
    "show_line_numbers": false
  },
  "results": {
    "page_size": 500,
    "max_column_width": 80
  },
  "connections": [
    {
      "name": "local pg",
      "dsn": "postgres://admin:secret@db.example.com:5432/production"
    },
    {
      "dsn": "./test.db"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.KeyMode != "vim" {
		t.Errorf("KeyMode = %q, want %q", cfg.KeyMode, "vim")
	}
	if cfg.Editor.TabSize != 2 {
		t.Errorf("Editor.TabSize = %d, want %d", cfg.Editor.TabSize, 2)
	}
	if cfg.Editor.ShowLineNumbers != false {
		t.Errorf("Editor.ShowLineNumbers = %v, want false", cfg.Editor.ShowLineNumbers)
	}
	if cfg.Results.PageSize != 500 {
		t.Errorf("Results.PageSize = %d, want %d", cfg.Results.PageSize, 500)
	}
	if cfg.Results.MaxColumnWidth != 80 {
		t.Errorf("Results.MaxColumnWidth = %d, want %d", cfg.Results.MaxColumnWidth, 80)
	}
	if len(cfg.Connections) != 2 {
		t.Fatalf("Connections length = %d, want 2", len(cfg.Connections))
	}

	c := cfg.Connections[0]
	if c.Name != "local pg" {
		t.Errorf("Connection[0].Name = %q, want %q", c.Name, "local pg")
	}
	if c.DSN != "postgres://admin:secret@db.example.com:5432/production" {
		t.Errorf("Connection[0].DSN = %q, want postgres DSN", c.DSN)
	}

	c2 := cfg.Connections[1]
	if c2.Name != "" {
		t.Errorf("Connection[1].Name = %q, want empty", c2.Name)
	}
	if c2.DSN != "./test.db" {
		t.Errorf("Connection[1].DSN = %q, want %q", c2.DSN, "./test.db")
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for missing file", err)
	}

	def := DefaultConfig()
	if !reflect.DeepEqual(cfg, def) {
		t.Errorf("Load(missing) = %+v, want DefaultConfig %+v", cfg, def)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	content := `{"keymode": "vim", invalid}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load(invalid JSON) error = nil, want error")
	}
}

func TestLoadPartialJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.json")

	data := `{"editor": {"tab_size": 8}}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Editor.TabSize != 8 {
		t.Errorf("Editor.TabSize = %d, want %d", cfg.Editor.TabSize, 8)
	}
	// These should remain at default values
	if cfg.KeyMode != "vim" {
		t.Errorf("KeyMode = %q, want default %q", cfg.KeyMode, "vim")
	}
	if cfg.Editor.ShowLineNumbers != true {
		t.Errorf("Editor.ShowLineNumbers = %v, want default true", cfg.Editor.ShowLineNumbers)
	}
	if cfg.Results.PageSize != 1000 {
		t.Errorf("Results.PageSize = %d, want default %d", cfg.Results.PageSize, 1000)
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	original := &Config{
		KeyMode: "vim",
		Editor: EditorConfig{
			TabSize:         3,
			ShowLineNumbers: false,
		},
		Results: ResultsConfig{
			PageSize:       200,
			MaxColumnWidth: 100,
		},
		Connections: []SavedConnection{
			{
				Name: "prod-pg",
				DSN:  "postgres://appuser:p%40ss@db.prod.internal:5433/maindb",
			},
		},
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(original, loaded) {
		t.Errorf("roundtrip mismatch:\n  saved:  %+v\n  loaded: %+v", original, loaded)
	}
}

func TestSaveDefaultAndLoadDefault(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	cfg := &Config{
		KeyMode: "vim",
		Editor: EditorConfig{
			TabSize:         2,
			ShowLineNumbers: true,
		},
		Results: ResultsConfig{
			PageSize:       100,
			MaxColumnWidth: 40,
		},
	}

	if err := cfg.SaveDefault(); err != nil {
		t.Fatalf("SaveDefault() error = %v", err)
	}

	loaded, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}

	if loaded.KeyMode != cfg.KeyMode {
		t.Errorf("KeyMode = %q, want %q", loaded.KeyMode, cfg.KeyMode)
	}
	if loaded.Editor != cfg.Editor {
		t.Errorf("Editor = %+v, want %+v", loaded.Editor, cfg.Editor)
	}
	if loaded.Results != cfg.Results {
		t.Errorf("Results = %+v, want %+v", loaded.Results, cfg.Results)
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	if dir == "" {
		t.Fatal("ConfigDir() returned empty string")
	}
	if filepath.Base(dir) != "seeql" {
		t.Errorf("ConfigDir() base = %q, want %q", filepath.Base(dir), "seeql")
	}
}
