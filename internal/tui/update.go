package tui

import (
	"tui-sqlite/internal/tui/components"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case LoadTableDataMsg:
		if msg.Err == nil {
			m.Grid.SetData(msg.Headers, msg.Rows, msg.TotalRows)
			m.Grid.ActiveTable = msg.TableName
		}
		return m, nil

	case LoadSchemaMsg:
		if msg.Err == nil {
			m.Grid.SchemaCols = msg.Columns
			m.Grid.SchemaDDL = msg.DDL
			m.Grid.ActiveTable = msg.TableName
		}
		return m, nil

	case components.PageChangedMsg:
		return m, loadTableDataCmd(m.DB, m.Grid.ActiveTable, msg.Page, m.Grid.PageSize)

	case tea.KeyMsg:
		// If sidebar has active filter input, it intercepts keys first
		if m.ActiveTab == SidebarTab && m.Sidebar.FilterActive {
			var cmd tea.Cmd
			m.Sidebar, cmd = m.Sidebar.Update(msg)
			return m, cmd
		}

		// Global key handlers
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			switch m.ActiveTab {
			case SidebarTab:
				m.ActiveTab = GridTab
			case GridTab:
				m.ActiveTab = EditorTab
			case EditorTab:
				m.ActiveTab = SidebarTab
			}
			return m, nil
		case "shift+tab":
			switch m.ActiveTab {
			case SidebarTab:
				m.ActiveTab = EditorTab
			case GridTab:
				m.ActiveTab = SidebarTab
			case EditorTab:
				m.ActiveTab = GridTab
			}
			return m, nil
		case "enter":
			if m.ActiveTab == SidebarTab {
				// Load table data and switch to GridTab
				if len(m.Sidebar.Filtered) > 0 && m.Sidebar.ActiveIndex < len(m.Sidebar.Filtered) {
					selTable := m.Sidebar.Filtered[m.Sidebar.ActiveIndex].Name
					m.Grid.ActiveTable = selTable
					m.Grid.CurrentPage = 1
					m.Grid.ScrollOffset = 0
					m.Grid.SchemaMode = false
					m.ActiveTab = GridTab
					return m, loadTableDataCmd(m.DB, selTable, 1, m.Grid.PageSize)
				}
			}
		case "s":
			if m.ActiveTab == SidebarTab || m.ActiveTab == GridTab {
				var selTable string
				if m.ActiveTab == SidebarTab && len(m.Sidebar.Filtered) > 0 && m.Sidebar.ActiveIndex < len(m.Sidebar.Filtered) {
					selTable = m.Sidebar.Filtered[m.Sidebar.ActiveIndex].Name
				} else {
					selTable = m.Grid.ActiveTable
				}

				if selTable != "" {
					m.Grid.ActiveTable = selTable
					m.Grid.SchemaMode = !m.Grid.SchemaMode
					if m.Grid.SchemaMode {
						m.ActiveTab = GridTab
						return m, loadSchemaCmd(m.DB, selTable)
					} else {
						m.Grid.CurrentPage = 1
						m.Grid.ScrollOffset = 0
						return m, loadTableDataCmd(m.DB, selTable, 1, m.Grid.PageSize)
					}
				}
			}
		}
	}

	// Propagate updates to the active component
	switch m.ActiveTab {
	case SidebarTab:
		var cmd tea.Cmd
		m.Sidebar, cmd = m.Sidebar.Update(msg)
		cmds = append(cmds, cmd)
	case GridTab:
		var cmd tea.Cmd
		m.Grid, cmd = m.Grid.Update(msg)
		cmds = append(cmds, cmd)
	case EditorTab:
		var cmd tea.Cmd
		m.Editor, cmd = m.Editor.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
