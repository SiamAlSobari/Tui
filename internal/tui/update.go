package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
