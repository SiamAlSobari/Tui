package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	dbPath   string
	readOnly bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	mode := "read-write"
	if m.readOnly {
		mode = "read-only"
	}
	return fmt.Sprintf("TuiSqlite v1.0.0\nDatabase: %s (%s)\n\nPress 'q' to quit.\n", m.dbPath, mode)
}

func main() {
	dbFlag := flag.String("db", "", "Path to SQLite database file")
	roFlag := flag.Bool("ro", false, "Open database in read-only mode")
	flag.Parse()

	dbPath := *dbFlag
	if dbPath == "" && flag.NArg() > 0 {
		dbPath = flag.Arg(0)
	}

	if dbPath == "" {
		fmt.Println("Usage: tui-sqlite [-ro] <database-path>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	m := model{
		dbPath:   dbPath,
		readOnly: *roFlag,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
