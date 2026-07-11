package components

import (
	"fmt"
	"strings"
	"tui-sqlite/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SidebarModel struct {
	Tables       []db.TableInfo
	Filtered     []db.TableInfo
	ActiveIndex  int
	FilterInput  textinput.Model
	FilterActive bool
}

func NewSidebar() SidebarModel {
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	return SidebarModel{
		FilterInput:  ti,
		Tables:       []db.TableInfo{},
		Filtered:     []db.TableInfo{},
		ActiveIndex:  0,
		FilterActive: false,
	}
}

func (m *SidebarModel) SetTables(tables []db.TableInfo) {
	m.Tables = tables
	m.Filtered = tables
	m.ActiveIndex = 0
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.FilterActive {
			switch msg.String() {
			case "esc":
				m.FilterActive = false
				m.FilterInput.Blur()
				m.FilterInput.SetValue("")
				m.Filtered = m.Tables
				m.ActiveIndex = 0
				return m, nil
			case "enter":
				m.FilterActive = false
				m.FilterInput.Blur()
				return m, nil
			}

			m.FilterInput, cmd = m.FilterInput.Update(msg)
			val := m.FilterInput.Value()
			m.Filtered = nil
			for _, t := range m.Tables {
				if strings.Contains(strings.ToLower(t.Name), strings.ToLower(val)) {
					m.Filtered = append(m.Filtered, t)
				}
			}
			if m.ActiveIndex >= len(m.Filtered) {
				m.ActiveIndex = len(m.Filtered) - 1
			}
			if m.ActiveIndex < 0 {
				m.ActiveIndex = 0
			}
			return m, cmd
		}

		// When filter is not active
		switch msg.String() {
		case "/":
			m.FilterActive = true
			m.FilterInput.Focus()
			m.FilterInput.SetValue("")
			m.Filtered = m.Tables
			m.ActiveIndex = 0
			return m, nil
		case "j", "down":
			if len(m.Filtered) > 0 {
				m.ActiveIndex++
				if m.ActiveIndex >= len(m.Filtered) {
					m.ActiveIndex = len(m.Filtered) - 1
				}
			}
		case "k", "up":
			m.ActiveIndex--
			if m.ActiveIndex < 0 {
				m.ActiveIndex = 0
			}
		}
	}

	return m, nil
}

func (m SidebarModel) View() string {
	var s strings.Builder

	s.WriteString("Tables & Views\n\n")

	if len(m.Filtered) == 0 {
		s.WriteString("  (No tables found)\n")
	} else {
		for i, t := range m.Filtered {
			prefix := "  "
			if i == m.ActiveIndex {
				prefix = "> "
			}
			s.WriteString(fmt.Sprintf("%s%s (%d)\n", prefix, t.Name, t.RowCount))
		}
	}

	s.WriteString("\n")
	s.WriteString("Filter: " + m.FilterInput.View() + "\n")

	return s.String()
}
