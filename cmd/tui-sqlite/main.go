package main

import (
	"flag"
	"fmt"
	"os"

	"tui-sqlite/internal/db"
	"tui-sqlite/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dbFlag := flag.String("db", "", "Path to SQLite database file")
	roFlag := flag.Bool("ro", false, "Open database in read-only mode")
	flag.Parse()

	dbPath := *dbFlag
	if dbPath == "" && flag.NArg() > 0 {
		dbPath = flag.Arg(0)
	}

	if dbPath == "" || flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "Usage: tui-sqlite [-ro] <database-path>")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.CommandLine.SetOutput(os.Stderr)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 1. Open database connection
	client, err := db.OpenConnection(dbPath, *roFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// 2. Fetch tables list
	tables, err := db.ListTables(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing database tables: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize TUI model and populate tables
	m := tui.NewModel(client)
	m.Sidebar.SetTables(tables)
	var tableNames []string
	for _, t := range tables {
		tableNames = append(tableNames, t.Name)
	}
	m.Editor.TableNames = tableNames

	// 4. Run Bubble Tea Program with AltScreen enabled for full screen interactive UI
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
