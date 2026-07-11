package tui

import (
	"fmt"
	"tui-sqlite/internal/db"
	"tui-sqlite/internal/tui/components"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type ActiveTab int

const (
	SidebarTab ActiveTab = iota
	GridTab
	EditorTab
)

type Model struct {
	DB            *db.DBClient
	Width, Height int
	ActiveTab     ActiveTab
	Sidebar       components.SidebarModel
	Grid          components.GridModel
	Editor        components.EditorModel
	StatusMessage string
	Loading       bool
	Spinner       spinner.Model
}

func NewModel(client *db.DBClient) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{
		DB:            client,
		ActiveTab:     SidebarTab,
		Sidebar:       components.NewSidebar(),
		Grid:          components.NewGrid(),
		Editor:        components.NewEditor(),
		StatusMessage: "",
		Loading:       false,
		Spinner:       s,
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.Spinner.Tick)
	if len(m.Sidebar.Tables) > 0 {
		firstTable := m.Sidebar.Tables[0].Name
		cmds = append(cmds, loadTableDataCmd(m.DB, firstTable, 1, 50))
	}
	return tea.Batch(cmds...)
}

type LoadTableDataMsg struct {
	TableName string
	Headers   []string
	Rows      [][]string
	TotalRows int
	Err       error
}

type LoadSchemaMsg struct {
	TableName string
	Columns   []db.ColumnInfo
	DDL       string
	Err       error
}

func loadTableDataCmd(client *db.DBClient, tableName string, page, pageSize int) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return LoadTableDataMsg{TableName: tableName}
		}
		offset := (page - 1) * pageSize
		headers, rows, err := db.QueryTablePage(client, tableName, pageSize, offset)
		if err != nil {
			return LoadTableDataMsg{Err: err}
		}

		// List tables to get accurate row counts
		tables, err := db.ListTables(client)
		totalRows := 0
		if err == nil {
			for _, t := range tables {
				if t.Name == tableName {
					totalRows = t.RowCount
					break
				}
			}
		}

		return LoadTableDataMsg{
			TableName: tableName,
			Headers:   headers,
			Rows:      rows,
			TotalRows: totalRows,
		}
	}
}

func loadSchemaCmd(client *db.DBClient, tableName string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return LoadSchemaMsg{TableName: tableName}
		}
		cols, ddl, err := db.GetTableSchema(client, tableName)
		if err != nil {
			return LoadSchemaMsg{Err: err}
		}
		return LoadSchemaMsg{
			TableName: tableName,
			Columns:   cols,
			DDL:       ddl,
		}
	}
}

type RunQueryResultMsg struct {
	Headers []string
	Rows    [][]string
	Err     error
}

func runQueryCmd(client *db.DBClient, sqlQuery string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return RunQueryResultMsg{Err: fmt.Errorf("no database connection")}
		}
		cols, rows, err := db.ExecuteSQL(client, sqlQuery)
		if err != nil {
			return RunQueryResultMsg{Err: err}
		}
		return RunQueryResultMsg{
			Headers: cols,
			Rows:    rows,
		}
	}
}

type RefreshTableMsg struct {
	TableName string
}

func deleteRowCmd(client *db.DBClient, tableName string, headers []string, row []string) tea.Cmd {
	return func() tea.Msg {
		err := db.DeleteRow(client, tableName, headers, row)
		if err != nil {
			return components.StatusMsg{Message: "Delete failed: " + err.Error()}
		}
		return RefreshTableMsg{TableName: tableName}
	}
}

func createRowCmd(client *db.DBClient, tableName string, headers []string) tea.Cmd {
	return func() tea.Msg {
		err := db.CreateRow(client, tableName, headers)
		if err != nil {
			return components.StatusMsg{Message: "Create failed: " + err.Error()}
		}
		return RefreshTableMsg{TableName: tableName}
	}
}

func updateCellCmd(client *db.DBClient, tableName string, headers []string, row []string, colIndex int, newValue string) tea.Cmd {
	return func() tea.Msg {
		err := db.UpdateCell(client, tableName, headers, row, colIndex, newValue)
		if err != nil {
			return components.StatusMsg{Message: "Update failed: " + err.Error()}
		}
		return RefreshTableMsg{TableName: tableName}
	}
}
