package tui

import (
	"tui-sqlite/internal/tui/components"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case LoadTableDataMsg:
		m.Loading = false
		if msg.Err == nil {
			m.Grid.SetData(msg.Headers, msg.Rows, msg.TotalRows)
			m.Grid.ActiveTable = msg.TableName
		}
		return m, nil

	case LoadSchemaMsg:
		m.Loading = false
		if msg.Err == nil {
			m.Grid.SchemaCols = msg.Columns
			m.Grid.SchemaDDL = msg.DDL
			m.Grid.ActiveTable = msg.TableName
		}
		return m, nil

	case components.PageChangedMsg:
		m.Loading = true
		return m, loadTableDataCmd(m.DB, m.Grid.ActiveTable, msg.Page, m.Grid.PageSize)

	case components.RunQueryMsg:
		m.Loading = true
		m.StatusMessage = "Executing query..."
		return m, runQueryCmd(m.DB, msg.SQL)

	case RunQueryResultMsg:
		m.Loading = false
		if msg.Err != nil {
			m.StatusMessage = "SQL Error: " + msg.Err.Error()
		} else {
			m.Grid.SetData(msg.Headers, msg.Rows, len(msg.Rows))
			m.Grid.ActiveTable = "Custom Query"
			m.Grid.CurrentPage = 1
			m.Grid.ScrollOffset = 0
			m.Grid.SchemaMode = false
			m.StatusMessage = "Query executed successfully!"
		}
		return m, nil

	case components.StatusMsg:
		m.StatusMessage = msg.Message
		return m, nil

	case RefreshTableMsg:
		m.StatusMessage = "Database updated successfully"
		m.Loading = true
		return m, loadTableDataCmd(m.DB, msg.TableName, m.Grid.CurrentPage, m.Grid.PageSize)

	case components.DeleteRowMsg:
		if m.DB.ReadOnly {
			m.StatusMessage = "Error: Database is opened in read-only mode"
			return m, nil
		}
		if msg.RowIndex < 0 || msg.RowIndex >= len(m.Grid.Rows) {
			return m, nil
		}
		row := m.Grid.Rows[msg.RowIndex]
		m.Loading = true
		m.StatusMessage = "Deleting row..."
		return m, deleteRowCmd(m.DB, m.Grid.ActiveTable, m.Grid.Headers, row)

	case components.CreateRowMsg:
		if m.DB.ReadOnly {
			m.StatusMessage = "Error: Database is opened in read-only mode"
			return m, nil
		}
		m.Loading = true
		m.StatusMessage = "Creating row..."
		return m, createRowCmd(m.DB, m.Grid.ActiveTable, m.Grid.Headers)

	case components.UpdateCellMsg:
		if m.DB.ReadOnly {
			m.StatusMessage = "Error: Database is opened in read-only mode"
			return m, nil
		}
		if msg.RowIndex < 0 || msg.RowIndex >= len(m.Grid.Rows) {
			return m, nil
		}
		row := m.Grid.Rows[msg.RowIndex]
		m.Loading = true
		m.StatusMessage = "Updating cell..."
		return m, updateCellCmd(m.DB, m.Grid.ActiveTable, m.Grid.Headers, row, msg.ColIndex, msg.Value)

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
					m.Loading = true
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
					m.Loading = true
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
	m.Grid.Focused = (m.ActiveTab == GridTab)

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
