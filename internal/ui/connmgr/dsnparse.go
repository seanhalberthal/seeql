package connmgr

import (
	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

// ParsedDSN holds extracted display info from a DSN string.
// This is for UI display only — the raw DSN is always passed to the adapter.
type ParsedDSN struct {
	Adapter  string            // "postgres", "mysql", "sqlite", ""
	Host     string            // hostname or file path
	Port     string            // port if present
	Database string            // database name
	User     string            // username (no password)
	Params   map[string]string // query parameters
	Valid    bool              // whether parsing succeeded
}

// ParseDSN extracts display info from a DSN string.
func ParseDSN(dsn string) ParsedDSN {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return ParsedDSN{}
	}

	lower := strings.ToLower(dsn)

	// SQLite: file paths and special schemes
	switch {
	case strings.HasPrefix(lower, "sqlite://"),
		strings.HasPrefix(lower, "file:"),
		strings.HasSuffix(lower, ".db"),
		strings.HasSuffix(lower, ".sqlite"),
		strings.HasSuffix(lower, ".sqlite3"),
		dsn == ":memory:":
		return parseSQLiteDSN(dsn)
	}

	// MySQL go-sql-driver format: user:pass@tcp(host:port)/db
	if strings.Contains(lower, "@tcp(") {
		return parseMySQLDriverDSN(dsn)
	}

	// URL-style: postgres://, postgresql://, mysql://
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		return parseURLDSN(u)
	}

	// PostgreSQL keyword format: host=localhost dbname=mydb
	if strings.Contains(dsn, "=") && !strings.Contains(dsn, "://") {
		return parsePGKeywordDSN(dsn)
	}

	return ParsedDSN{}
}

func parseURLDSN(u *url.URL) ParsedDSN {
	p := ParsedDSN{Valid: true}

	switch strings.ToLower(u.Scheme) {
	case "postgres", "postgresql":
		p.Adapter = "postgres"
	case "mysql":
		p.Adapter = "mysql"
	case "sqlite", "file":
		p.Adapter = "sqlite"
		p.Database = strings.TrimPrefix(u.Path, "/")
		if p.Database == "" {
			p.Database = u.Opaque
		}
		return p
	default:
		return ParsedDSN{}
	}

	p.Host = u.Hostname()
	p.Port = u.Port()
	p.Database = strings.TrimPrefix(u.Path, "/")
	if u.User != nil {
		p.User = u.User.Username()
	}
	p.Params = make(map[string]string)
	for k, v := range u.Query() {
		if len(v) > 0 {
			p.Params[k] = v[0]
		}
	}
	return p
}

func parseSQLiteDSN(dsn string) ParsedDSN {
	path := dsn
	for _, prefix := range []string{"sqlite://", "file:"} {
		path = strings.TrimPrefix(path, prefix)
	}
	return ParsedDSN{
		Adapter:  "sqlite",
		Database: filepath.Base(path),
		Host:     path,
		Valid:    true,
	}
}

func parseMySQLDriverDSN(dsn string) ParsedDSN {
	p := ParsedDSN{Adapter: "mysql", Valid: true, Params: make(map[string]string)}

	atIdx := strings.Index(dsn, "@tcp(")
	if atIdx < 0 {
		return ParsedDSN{}
	}

	userPart := dsn[:atIdx]
	if colonIdx := strings.Index(userPart, ":"); colonIdx >= 0 {
		p.User = userPart[:colonIdx]
	} else {
		p.User = userPart
	}

	rest := dsn[atIdx+5:] // skip "@tcp("
	closeIdx := strings.Index(rest, ")")
	if closeIdx < 0 {
		return ParsedDSN{}
	}

	hostPort := rest[:closeIdx]
	if colonIdx := strings.LastIndex(hostPort, ":"); colonIdx >= 0 {
		p.Host = hostPort[:colonIdx]
		p.Port = hostPort[colonIdx+1:]
	} else {
		p.Host = hostPort
	}

	rest = rest[closeIdx+1:]
	if qIdx := strings.Index(rest, "?"); qIdx >= 0 {
		p.Database = strings.TrimPrefix(rest[:qIdx], "/")
		if u, err := url.ParseQuery(rest[qIdx+1:]); err == nil {
			for k, v := range u {
				if len(v) > 0 {
					p.Params[k] = v[0]
				}
			}
		}
	} else {
		p.Database = strings.TrimPrefix(rest, "/")
	}

	return p
}

func parsePGKeywordDSN(dsn string) ParsedDSN {
	p := ParsedDSN{Adapter: "postgres", Valid: true, Params: make(map[string]string)}
	for _, part := range strings.Fields(dsn) {
		eqIdx := strings.Index(part, "=")
		if eqIdx < 0 {
			continue
		}
		key := part[:eqIdx]
		val := part[eqIdx+1:]
		switch key {
		case "host":
			p.Host = val
		case "port":
			p.Port = val
		case "dbname":
			p.Database = val
		case "user":
			p.User = val
		case "password":
			// Don't store
		default:
			p.Params[key] = val
		}
	}
	return p
}

// Summary returns a short display string like "postgres · localhost · mydb".
func (p ParsedDSN) Summary() string {
	if !p.Valid {
		return ""
	}
	parts := []string{p.Adapter}
	if p.Host != "" {
		parts = append(parts, p.Host)
	}
	if p.Database != "" {
		parts = append(parts, p.Database)
	}
	return strings.Join(parts, " \u00b7 ")
}

// ParamString returns query params as a display string like "sslmode=disable, connect_timeout=10".
func (p ParsedDSN) ParamString() string {
	if len(p.Params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(p.Params))
	for k := range p.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + p.Params[k]
	}
	return strings.Join(parts, ", ")
}
