package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type RunQueryMsg struct {
	SQL string
}

type EditorModel struct {
	Input      textarea.Model
	History    []string
	HistoryIdx int
}

func NewEditor() EditorModel {
	ta := textarea.New()
	ta.Placeholder = "SELECT * FROM users;\n(Press Ctrl+J or Ctrl+Enter to execute)"
	ta.Focus()
	// Set initial size
	ta.SetWidth(60)
	ta.SetHeight(8)
	return EditorModel{
		Input:      ta,
		History:    []string{},
		HistoryIdx: -1,
	}
}

func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d", "ctrl+enter", "ctrl+j":
			sqlText := m.Input.Value()
			if strings.TrimSpace(sqlText) != "" {
				// Save to history
				if len(m.History) == 0 || m.History[len(m.History)-1] != sqlText {
					m.History = append(m.History, sqlText)
				}
				m.HistoryIdx = -1
				return m, func() tea.Msg {
					return RunQueryMsg{SQL: sqlText}
				}
			}

		case "up":
			// If cursor is at the first line, navigate history
			if m.Input.Line() == 0 && len(m.History) > 0 {
				if m.HistoryIdx == -1 {
					m.HistoryIdx = len(m.History) - 1
				} else if m.HistoryIdx > 0 {
					m.HistoryIdx--
				}
				m.Input.SetValue(m.History[m.HistoryIdx])
				// Move cursor to end
				m.Input.SetCursor(len(m.Input.Value()))
				return m, nil
			}
			// Otherwise fallback to default textarea handling
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd

		case "down":
			// If cursor is at the last line, navigate history
			lines := strings.Split(m.Input.Value(), "\n")
			if m.Input.Line() == len(lines)-1 && len(m.History) > 0 {
				if m.HistoryIdx != -1 {
					if m.HistoryIdx < len(m.History)-1 {
						m.HistoryIdx++
						m.Input.SetValue(m.History[m.HistoryIdx])
					} else {
						m.HistoryIdx = -1
						m.Input.SetValue("")
					}
					m.Input.SetCursor(len(m.Input.Value()))
					return m, nil
				}
			}
			// Otherwise fallback to default textarea handling
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd
		}
	}

	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m EditorModel) View() string {
	var s strings.Builder
	s.WriteString("SQL QUERY EDITOR\n")
	s.WriteString("Write custom SQL statements here (Ctrl+J or Ctrl+Enter to Run):\n\n")
	s.WriteString(m.Input.View())
	s.WriteString("\n")
	return s.String()
}
