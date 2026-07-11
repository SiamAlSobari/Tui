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
	return nil
}
