package tui

import (
	"tui-sqlite/internal/db"
	"tui-sqlite/internal/tui/components"

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
}

func NewModel(client *db.DBClient) Model {
	return Model{
		DB:        client,
		ActiveTab: SidebarTab,
		Sidebar:   components.NewSidebar(),
		Grid:      components.NewGrid(),
		Editor:    components.NewEditor(),
	}
}

func (m Model) Init() tea.Cmd {
	if len(m.Sidebar.Tables) > 0 {
		firstTable := m.Sidebar.Tables[0].Name
		return loadTableDataCmd(m.DB, firstTable, 1, 50)
	}
	return nil
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
