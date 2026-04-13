package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
)

// mockAdapter is a minimal adapter for testing the registry.
type mockAdapter struct {
	name string
	port int
}

func (m *mockAdapter) Name() string     { return m.name }
func (m *mockAdapter) DefaultPort() int { return m.port }
func (m *mockAdapter) Connect(_ context.Context, _ string) (Connection, error) {
	return nil, errors.New("mock: not implemented")
}

func TestRegister(t *testing.T) {
	// Save and restore original registry state.
	orig := make(map[string]Adapter)
	for k, v := range Registry {
		orig[k] = v
	}
	defer func() {
		Registry = orig
	}()

	// Clear registry for this test.
	Registry = map[string]Adapter{}

	mock := &mockAdapter{name: "testdb", port: 9999}
	Register(mock)

	got, ok := Registry["testdb"]
	if !ok {
		t.Fatal("expected adapter 'testdb' to be registered")
	}
	if got.Name() != "testdb" {
		t.Errorf("Name() = %q, want %q", got.Name(), "testdb")
	}
	if got.DefaultPort() != 9999 {
		t.Errorf("DefaultPort() = %d, want %d", got.DefaultPort(), 9999)
	}
}

func TestRegister_Multiple(t *testing.T) {
	orig := make(map[string]Adapter)
	for k, v := range Registry {
		orig[k] = v
	}
	defer func() {
		Registry = orig
	}()

	Registry = map[string]Adapter{}

	adapters := []struct {
		name string
		port int
	}{
		{"alpha", 1111},
		{"bravo", 2222},
		{"charlie", 3333},
	}

	for _, a := range adapters {
		Register(&mockAdapter{name: a.name, port: a.port})
	}

	if len(Registry) != 3 {
		t.Fatalf("expected 3 adapters in registry, got %d", len(Registry))
	}

	for _, a := range adapters {
		got, ok := Registry[a.name]
		if !ok {
			t.Errorf("adapter %q not found in registry", a.name)
			continue
		}
		if got.Name() != a.name {
			t.Errorf("Name() = %q, want %q", got.Name(), a.name)
		}
		if got.DefaultPort() != a.port {
			t.Errorf("DefaultPort() for %q = %d, want %d", a.name, got.DefaultPort(), a.port)
		}
	}
}

func TestSentinelEOF(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"io.EOF returns true", io.EOF, true},
		{"nil returns false", nil, false},
		{"other error returns false", errors.New("some error"), false},
		{"wrapped io.EOF returns true", fmt.Errorf("wrap: %w", io.EOF), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SentinelEOF(tt.err); got != tt.want {
				t.Errorf("SentinelEOF(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestQueryResult_IsSelect(t *testing.T) {
	selectResult := &QueryResult{
		IsSelect: true,
		RowCount: 5,
		Columns:  []ColumnMeta{{Name: "id", Type: "int4"}},
	}
	if !selectResult.IsSelect {
		t.Error("expected IsSelect to be true")
	}

	nonSelectResult := &QueryResult{
		IsSelect: false,
		RowCount: 1,
		Message:  "INSERT 0 1",
	}
	if nonSelectResult.IsSelect {
		t.Error("expected IsSelect to be false")
	}
}

func TestColumnMeta(t *testing.T) {
	col := ColumnMeta{
		Name:     "user_id",
		Type:     "int4",
		Nullable: true,
	}

	if col.Name != "user_id" {
		t.Errorf("Name = %q, want %q", col.Name, "user_id")
	}
	if col.Type != "int4" {
		t.Errorf("Type = %q, want %q", col.Type, "int4")
	}
	if !col.Nullable {
		t.Error("expected Nullable to be true")
	}

	nonNullCol := ColumnMeta{
		Name:     "email",
		Type:     "varchar",
		Nullable: false,
	}
	if nonNullCol.Nullable {
		t.Error("expected Nullable to be false")
	}
}

func TestCompletionKind(t *testing.T) {
	kinds := []CompletionKind{
		CompletionTable,
		CompletionColumn,
		CompletionKeyword,
		CompletionFunction,
		CompletionSchema,
		CompletionDatabase,
		CompletionView,
	}

	// All values must be distinct.
	seen := make(map[CompletionKind]string)
	names := []string{
		"CompletionTable",
		"CompletionColumn",
		"CompletionKeyword",
		"CompletionFunction",
		"CompletionSchema",
		"CompletionDatabase",
		"CompletionView",
	}

	for i, k := range kinds {
		if existing, ok := seen[k]; ok {
			t.Errorf("%s has the same value (%d) as %s", names[i], k, existing)
		}
		seen[k] = names[i]
	}

	// Verify the iota-based ordering starts at 0.
	if CompletionTable != 0 {
		t.Errorf("CompletionTable = %d, want 0", CompletionTable)
	}
	if CompletionView != 6 {
		t.Errorf("CompletionView = %d, want 6", CompletionView)
	}
}

func TestErrors(t *testing.T) {
	// All sentinel errors must be non-nil.
	if ErrNoBidirectional == nil {
		t.Error("ErrNoBidirectional is nil")
	}
	if ErrNotConnected == nil {
		t.Error("ErrNotConnected is nil")
	}
	if ErrCancelled == nil {
		t.Error("ErrCancelled is nil")
	}

	// All sentinel errors must be distinct.
	if errors.Is(ErrNoBidirectional, ErrNotConnected) {
		t.Error("ErrNoBidirectional and ErrNotConnected should be distinct")
	}
	if errors.Is(ErrNoBidirectional, ErrCancelled) {
		t.Error("ErrNoBidirectional and ErrCancelled should be distinct")
	}
	if errors.Is(ErrNotConnected, ErrCancelled) {
		t.Error("ErrNotConnected and ErrCancelled should be distinct")
	}

	// Verify error messages are non-empty and distinct.
	msgs := map[string]bool{
		ErrNoBidirectional.Error(): true,
		ErrNotConnected.Error():    true,
		ErrCancelled.Error():       true,
	}
	if len(msgs) != 3 {
		t.Error("expected 3 distinct error messages")
	}
}

func TestDetectAdapter(t *testing.T) {
	tests := []struct {
		dsn  string
		want string
	}{
		{"postgres://user:pass@host/db", "postgres"},
		{"postgresql://user@host/db", "postgres"},
		{"postgres://user:pass@host/db?sslmode=disable", "postgres"},
		{"postgres://user:pass@host/db?sslmode=disable&connect_timeout=10", "postgres"},
		{"mysql://root:pass@host/db", "mysql"},
		{"mysql://root:pass@host/db?charset=utf8mb4&tls=true", "mysql"},
		{"root:pass@tcp(host:3306)/db", "mysql"},
		{"root:pass@tcp(host:3306)/db?parseTime=true", "mysql"},
		{"sqlite:///tmp/test.db", "sqlite"},
		{"file:test.db", "sqlite"},
		{"test.db", "sqlite"},
		{"test.sqlite", "sqlite"},
		{"test.sqlite3", "sqlite"},
		{"user@host", "postgres"}, // fallback: contains @
		{"not-a-dsn", ""},
	}
	for _, tt := range tests {
		got := DetectAdapter(tt.dsn)
		if got != tt.want {
			t.Errorf("DetectAdapter(%q) = %q, want %q", tt.dsn, got, tt.want)
		}
	}
}
