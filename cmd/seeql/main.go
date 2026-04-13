package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/seanhalberthal/seeql/internal/adapter"
	"github.com/seanhalberthal/seeql/internal/app"
	"github.com/seanhalberthal/seeql/internal/audit"
	"github.com/seanhalberthal/seeql/internal/config"
	"github.com/seanhalberthal/seeql/internal/history"

	// Register database adapters
	_ "github.com/seanhalberthal/seeql/internal/adapter/mysql"
	_ "github.com/seanhalberthal/seeql/internal/adapter/postgres"
	_ "github.com/seanhalberthal/seeql/internal/adapter/sqlite"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var configFlag string

	rootCmd := &cobra.Command{
		Use:   "seeql [dsn]",
		Short: "A terminal SQL client",
		Long: `seeql is a terminal SQL client supporting PostgreSQL, MySQL, and SQLite.

Examples:
  seeql                                        # Launch connection manager
  seeql "postgres://user:pass@host/db"         # Connect via DSN
  seeql "postgres://user@host/db?sslmode=disable"  # DSN with query params (quote the URL!)
  seeql ./data.db                              # SQLite file
  seeql "mysql://user:pass@tcp(host)/db"       # MySQL connection`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			var cfg *config.Config
			var err error
			if configFlag != "" {
				cfg, err = config.Load(configFlag)
			} else {
				cfg, err = config.LoadDefault()
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
				cfg = config.DefaultConfig()
			}

			// Open history
			hist, err := history.New()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not open history: %v\n", err)
			}
			if hist != nil {
				defer hist.Close()
			}

			// Open audit log
			var auditLog *audit.Logger
			if cfg.Audit.Enabled {
				auditPath := cfg.Audit.Path
				if auditPath == "" {
					if dir, err := config.ConfigDir(); err == nil {
						auditPath = dir + "/audit.jsonl"
					}
				}
				if auditPath != "" {
					auditLog, err = audit.New(auditPath, cfg.Audit.MaxSizeMB)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not open audit log: %v\n", err)
					}
				}
			}
			if auditLog != nil {
				defer auditLog.Close()
			}

			// Create app model
			model := app.New(cfg, hist, auditLog)
			model.SetVersion(version)

			// Determine connection method
			var dsn string
			var adapterName string

			if len(args) > 0 {
				dsn = args[0]
				adapterName = adapter.DetectAdapter(dsn)
			}

			// If we have connection info, connect; otherwise show connection manager
			var initCmd tea.Cmd
			if adapterName != "" && dsn != "" {
				if _, ok := adapter.Registry[adapterName]; !ok {
					return fmt.Errorf("unknown adapter: %s (available: %s)", adapterName, availableAdapters())
				}
				initCmd = model.InitialConnect(adapterName, dsn)
			} else {
				model.ShowConnManager()
			}

			// Run the TUI
			p := tea.NewProgram(
				model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			if initCmd != nil {
				go func() {
					p.Send(initCmd())
				}()
			}

			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("error running application: %w", err)
			}

			// Close database connection if open
			if m, ok := finalModel.(app.Model); ok {
				if conn := m.Connection(); conn != nil {
					_ = conn.Close()
				}
			}

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&configFlag, "config", "c", "", "Config file path")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("seeql %s (commit: %s, built: %s)\n", version, commit, date)
			fmt.Println("\nSupported adapters:")
			for name := range adapter.Registry {
				fmt.Printf("  - %s\n", name)
			}
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func availableAdapters() string {
	var names []string
	for name := range adapter.Registry {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
