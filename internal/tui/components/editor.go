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
	TableNames []string
}

func NewEditor() EditorModel {
	ta := textarea.New()
	ta.Placeholder = "SELECT * FROM users;\n(Press Ctrl+J or Ctrl+Enter to execute)"
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(8)
	return EditorModel{
		Input:      ta,
		History:    []string{},
		HistoryIdx: -1,
		TableNames: []string{},
	}
}

var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "INSERT", "INTO", "UPDATE", "SET", "DELETE",
	"CREATE", "TABLE", "INDEX", "DROP", "JOIN", "ON", "GROUP BY", "ORDER BY",
	"LIMIT", "OFFSET", "AND", "OR", "NOT", "LIKE", "IN", "IS", "NULL",
	"PRAGMA", "EXPLAIN", "QUERY", "PLAN",
}

func getWordUnderCursor(val string, cursor int) string {
	if cursor < 0 || cursor > len(val) {
		return ""
	}
	start := cursor
	for start > 0 {
		r := rune(val[start-1])
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			start--
		} else {
			break
		}
	}
	return val[start:cursor]
}

func (m EditorModel) getCursorOffset() int {
	val := m.Input.Value()
	lines := strings.Split(val, "\n")
	currentLine := m.Input.Line()
	offset := 0
	for i := 0; i < currentLine && i < len(lines); i++ {
		offset += len(lines[i]) + 1
	}
	offset += m.Input.LineInfo().ColumnOffset
	if offset > len(val) {
		offset = len(val)
	}
	return offset
}

func (m EditorModel) getSuggestions(word string) []string {
	if word == "" {
		return nil
	}
	lowerWord := strings.ToLower(word)
	var matches []string

	// Check table names first
	for _, t := range m.TableNames {
		if strings.HasPrefix(strings.ToLower(t), lowerWord) {
			matches = append(matches, t)
		}
	}

	// Check SQL keywords
	for _, kw := range sqlKeywords {
		if strings.HasPrefix(strings.ToLower(kw), lowerWord) {
			matches = append(matches, kw)
		}
	}

	return matches
}

func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d", "ctrl+enter", "ctrl+j":
			sqlText := m.Input.Value()
			if strings.TrimSpace(sqlText) != "" {
				if len(m.History) == 0 || m.History[len(m.History)-1] != sqlText {
					m.History = append(m.History, sqlText)
				}
				m.HistoryIdx = -1
				return m, func() tea.Msg {
					return RunQueryMsg{SQL: sqlText}
				}
			}

		case "ctrl+space", "ctrl+l":
			val := m.Input.Value()
			cur := m.getCursorOffset()
			word := getWordUnderCursor(val, cur)
			if word != "" {
				sugs := m.getSuggestions(word)
				if len(sugs) > 0 {
					suggestion := sugs[0]
					startIdx := cur - len(word)
					newVal := val[:startIdx] + suggestion + val[cur:]
					m.Input.SetValue(newVal)
					m.Input.SetCursor(startIdx + len(suggestion))
				}
			}
			return m, nil

		case "up":
			if m.Input.Line() == 0 && len(m.History) > 0 {
				if m.HistoryIdx == -1 {
					m.HistoryIdx = len(m.History) - 1
				} else if m.HistoryIdx > 0 {
					m.HistoryIdx--
				}
				m.Input.SetValue(m.History[m.HistoryIdx])
				m.Input.SetCursor(len(m.Input.Value()))
				return m, nil
			}
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd

		case "down":
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

	// Render Autocomplete Suggestions
	val := m.Input.Value()
	cur := m.getCursorOffset()
	word := getWordUnderCursor(val, cur)
	if word != "" {
		sugs := m.getSuggestions(word)
		if len(sugs) > 0 {
			limit := 5
			if len(sugs) < limit {
				limit = len(sugs)
			}
			s.WriteString(" Suggestions: " + strings.Join(sugs[:limit], ", "))
			if len(sugs) > limit {
				s.WriteString("...")
			}
			s.WriteString(" (Press Ctrl+Space or Ctrl+L to complete)\n")
		}
	}
	return s.String()
}
