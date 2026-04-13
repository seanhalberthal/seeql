package connmgr

import "testing"

func TestParseDSN_Empty(t *testing.T) {
	p := ParseDSN("")
	if p.Valid {
		t.Fatal("expected invalid for empty DSN")
	}
}

func TestParseDSN_Whitespace(t *testing.T) {
	p := ParseDSN("   ")
	if p.Valid {
		t.Fatal("expected invalid for whitespace DSN")
	}
}

func TestParseDSN_Garbage(t *testing.T) {
	p := ParseDSN("not a real dsn at all")
	if p.Valid {
		t.Fatal("expected invalid for garbage input")
	}
}

func TestParseDSN_PostgresURL(t *testing.T) {
	p := ParseDSN("postgres://admin:secret@localhost:5432/mydb")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "postgres" {
		t.Fatalf("expected adapter=postgres, got %q", p.Adapter)
	}
	if p.Host != "localhost" {
		t.Fatalf("expected host=localhost, got %q", p.Host)
	}
	if p.Port != "5432" {
		t.Fatalf("expected port=5432, got %q", p.Port)
	}
	if p.Database != "mydb" {
		t.Fatalf("expected database=mydb, got %q", p.Database)
	}
	if p.User != "admin" {
		t.Fatalf("expected user=admin, got %q", p.User)
	}
}

func TestParseDSN_PostgresURL_WithParams(t *testing.T) {
	p := ParseDSN("postgres://user:pass@localhost:5432/mydb?sslmode=disable&connect_timeout=10")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "postgres" {
		t.Fatalf("expected adapter=postgres, got %q", p.Adapter)
	}
	if p.Database != "mydb" {
		t.Fatalf("expected database=mydb, got %q", p.Database)
	}
	if p.Params["sslmode"] != "disable" {
		t.Fatalf("expected sslmode=disable, got %q", p.Params["sslmode"])
	}
	if p.Params["connect_timeout"] != "10" {
		t.Fatalf("expected connect_timeout=10, got %q", p.Params["connect_timeout"])
	}
}

func TestParseDSN_PostgresqlScheme(t *testing.T) {
	p := ParseDSN("postgresql://user@host:5433/db")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "postgres" {
		t.Fatalf("expected adapter=postgres, got %q", p.Adapter)
	}
	if p.Host != "host" {
		t.Fatalf("expected host=host, got %q", p.Host)
	}
	if p.Port != "5433" {
		t.Fatalf("expected port=5433, got %q", p.Port)
	}
}

func TestParseDSN_MySQLURL(t *testing.T) {
	p := ParseDSN("mysql://root:pass@db.example.com:3306/mydb")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "mysql" {
		t.Fatalf("expected adapter=mysql, got %q", p.Adapter)
	}
	if p.Host != "db.example.com" {
		t.Fatalf("expected host=db.example.com, got %q", p.Host)
	}
	if p.Port != "3306" {
		t.Fatalf("expected port=3306, got %q", p.Port)
	}
	if p.Database != "mydb" {
		t.Fatalf("expected database=mydb, got %q", p.Database)
	}
}

func TestParseDSN_MySQLURL_WithParams(t *testing.T) {
	p := ParseDSN("mysql://root:pass@localhost:3306/mydb?charset=utf8mb4&tls=true")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Params["charset"] != "utf8mb4" {
		t.Fatalf("expected charset=utf8mb4, got %q", p.Params["charset"])
	}
	if p.Params["tls"] != "true" {
		t.Fatalf("expected tls=true, got %q", p.Params["tls"])
	}
}

func TestParseDSN_MySQLDriver(t *testing.T) {
	p := ParseDSN("root:password@tcp(localhost:3306)/mydb")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "mysql" {
		t.Fatalf("expected adapter=mysql, got %q", p.Adapter)
	}
	if p.Host != "localhost" {
		t.Fatalf("expected host=localhost, got %q", p.Host)
	}
	if p.Port != "3306" {
		t.Fatalf("expected port=3306, got %q", p.Port)
	}
	if p.Database != "mydb" {
		t.Fatalf("expected database=mydb, got %q", p.Database)
	}
	if p.User != "root" {
		t.Fatalf("expected user=root, got %q", p.User)
	}
}

func TestParseDSN_MySQLDriver_WithParams(t *testing.T) {
	p := ParseDSN("root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Params["charset"] != "utf8mb4" {
		t.Fatalf("expected charset=utf8mb4, got %q", p.Params["charset"])
	}
	if p.Params["parseTime"] != "true" {
		t.Fatalf("expected parseTime=true, got %q", p.Params["parseTime"])
	}
}

func TestParseDSN_MySQLDriver_NoPassword(t *testing.T) {
	p := ParseDSN("root@tcp(localhost:3306)/mydb")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.User != "root" {
		t.Fatalf("expected user=root, got %q", p.User)
	}
}

func TestParseDSN_SQLitePath(t *testing.T) {
	p := ParseDSN("/path/to/data.db")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "sqlite" {
		t.Fatalf("expected adapter=sqlite, got %q", p.Adapter)
	}
	if p.Database != "data.db" {
		t.Fatalf("expected database=data.db, got %q", p.Database)
	}
	if p.Host != "/path/to/data.db" {
		t.Fatalf("expected host=/path/to/data.db, got %q", p.Host)
	}
}

func TestParseDSN_SQLiteScheme(t *testing.T) {
	p := ParseDSN("sqlite:///tmp/test.db")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "sqlite" {
		t.Fatalf("expected adapter=sqlite, got %q", p.Adapter)
	}
}

func TestParseDSN_SQLiteFile(t *testing.T) {
	p := ParseDSN("file:test.db")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "sqlite" {
		t.Fatalf("expected adapter=sqlite, got %q", p.Adapter)
	}
}

func TestParseDSN_SQLiteMemory(t *testing.T) {
	p := ParseDSN(":memory:")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "sqlite" {
		t.Fatalf("expected adapter=sqlite, got %q", p.Adapter)
	}
}

func TestParseDSN_SQLiteSuffix(t *testing.T) {
	for _, suffix := range []string{".db", ".sqlite", ".sqlite3"} {
		p := ParseDSN("test" + suffix)
		if !p.Valid {
			t.Fatalf("expected valid for suffix %q", suffix)
		}
		if p.Adapter != "sqlite" {
			t.Fatalf("expected adapter=sqlite for suffix %q, got %q", suffix, p.Adapter)
		}
	}
}

func TestParseDSN_PGKeyword(t *testing.T) {
	p := ParseDSN("host=localhost port=5432 dbname=mydb user=admin password=secret sslmode=disable")
	if !p.Valid {
		t.Fatal("expected valid")
	}
	if p.Adapter != "postgres" {
		t.Fatalf("expected adapter=postgres, got %q", p.Adapter)
	}
	if p.Host != "localhost" {
		t.Fatalf("expected host=localhost, got %q", p.Host)
	}
	if p.Port != "5432" {
		t.Fatalf("expected port=5432, got %q", p.Port)
	}
	if p.Database != "mydb" {
		t.Fatalf("expected database=mydb, got %q", p.Database)
	}
	if p.User != "admin" {
		t.Fatalf("expected user=admin, got %q", p.User)
	}
	if p.Params["sslmode"] != "disable" {
		t.Fatalf("expected sslmode=disable, got %q", p.Params["sslmode"])
	}
}

func TestParseDSN_PGKeyword_PasswordOmitted(t *testing.T) {
	p := ParseDSN("host=localhost dbname=mydb user=admin password=supersecret")
	// Password should not appear in parsed output
	if _, ok := p.Params["password"]; ok {
		t.Fatal("password should not be stored in Params")
	}
}

func TestParsedDSN_Summary(t *testing.T) {
	tests := []struct {
		name string
		p    ParsedDSN
		want string
	}{
		{
			name: "invalid",
			p:    ParsedDSN{},
			want: "",
		},
		{
			name: "postgres full",
			p:    ParsedDSN{Valid: true, Adapter: "postgres", Host: "localhost", Database: "mydb"},
			want: "postgres \u00b7 localhost \u00b7 mydb",
		},
		{
			name: "sqlite no host",
			p:    ParsedDSN{Valid: true, Adapter: "sqlite", Database: "test.db"},
			want: "sqlite \u00b7 test.db",
		},
		{
			name: "adapter only",
			p:    ParsedDSN{Valid: true, Adapter: "mysql"},
			want: "mysql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Summary()
			if got != tt.want {
				t.Fatalf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsedDSN_ParamString(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "empty",
			params: nil,
			want:   "",
		},
		{
			name:   "single param",
			params: map[string]string{"sslmode": "disable"},
			want:   "sslmode=disable",
		},
		{
			name:   "multiple params sorted",
			params: map[string]string{"sslmode": "disable", "connect_timeout": "10"},
			want:   "connect_timeout=10, sslmode=disable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsedDSN{Valid: true, Params: tt.params}
			got := p.ParamString()
			if got != tt.want {
				t.Fatalf("ParamString() = %q, want %q", got, tt.want)
			}
		})
	}
}
