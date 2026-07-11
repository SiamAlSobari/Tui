package components

import (
	"fmt"
	"strings"
	"dbbee/internal/db"
	"dbbee/internal/export"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PageChangedMsg struct {
	Page int
}

type StatusMsg struct {
	Message string
}

type DeleteRowMsg struct {
	RowIndex int
}

type CreateRowMsg struct{}

type UpdateCellMsg struct {
	RowIndex int
	ColIndex int
	Value    string
}

type GridModel struct {
	Headers       []string
	Rows          [][]string
	Width, Height int
	PageSize      int
	CurrentPage   int
	TotalRows     int
	ScrollOffset  int
	ColWidths     map[int]int
	ActiveTable   string

	// Schema Inspector fields
	SchemaMode bool
	SchemaCols []db.ColumnInfo
	SchemaDDL  string

	// Writable operations & Navigation fields
	ActiveRowIndex int
	ActiveColIndex int
	ConfirmDelete  bool
	Focused        bool
	EditMode       bool
	EditInput      textinput.Model
}

func NewGrid() GridModel {
	ti := textinput.New()
	ti.Placeholder = "New value..."
	return GridModel{
		Headers:        []string{},
		Rows:           [][]string{},
		PageSize:       50,
		CurrentPage:    1,
		ColWidths:      make(map[int]int),
		ScrollOffset:   0,
		SchemaMode:     false,
		ActiveRowIndex: 0,
		ActiveColIndex: 0,
		ConfirmDelete:  false,
		Focused:        false,
		EditMode:       false,
		EditInput:      ti,
	}
}

func (m *GridModel) SetData(headers []string, rows [][]string, totalRows int) {
	m.Headers = headers
	m.Rows = rows
	m.TotalRows = totalRows

	// Calculate column widths
	m.ColWidths = make(map[int]int)
	for i, h := range headers {
		m.ColWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, val := range row {
			if i < len(headers) {
				if len(val) > m.ColWidths[i] {
					m.ColWidths[i] = len(val)
				}
			}
		}
	}
}

func (m *GridModel) adjustScrollOffset() {
	if len(m.Headers) == 0 {
		return
	}
	if m.ActiveColIndex < m.ScrollOffset {
		m.ScrollOffset = m.ActiveColIndex
	}
	// Calculate visible columns starting from ScrollOffset
	currentWidth := 0
	visibleCount := 0
	for i := m.ScrollOffset; i < len(m.Headers); i++ {
		colW := m.ColWidths[i] + 2
		if m.Width > 0 && currentWidth+colW+1 > m.Width && visibleCount > 0 {
			break
		}
		visibleCount++
		currentWidth += colW + 1
	}
	if m.ActiveColIndex >= m.ScrollOffset+visibleCount {
		m.ScrollOffset = m.ActiveColIndex - visibleCount + 1
		if m.ScrollOffset < 0 {
			m.ScrollOffset = 0
		}
		if m.ScrollOffset >= len(m.Headers) {
			m.ScrollOffset = len(m.Headers) - 1
		}
	}
}

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.EditMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.EditMode = false
				val := m.EditInput.Value()
				m.EditInput.SetValue("")
				m.EditInput.Blur()
				return m, func() tea.Msg {
					return UpdateCellMsg{
						RowIndex: m.ActiveRowIndex,
						ColIndex: m.ActiveColIndex,
						Value:    val,
					}
				}
			case "esc":
				m.EditMode = false
				m.EditInput.SetValue("")
				m.EditInput.Blur()
				return m, nil
			}
		}
		m.EditInput, cmd = m.EditInput.Update(msg)
		return m, cmd
	}

	if m.ConfirmDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.ConfirmDelete = false
				return m, func() tea.Msg {
					return DeleteRowMsg{RowIndex: m.ActiveRowIndex}
				}
			default:
				m.ConfirmDelete = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l", "right":
			if m.ActiveColIndex < len(m.Headers)-1 {
				m.ActiveColIndex++
				m.adjustScrollOffset()
			}
		case "h", "left":
			if m.ActiveColIndex > 0 {
				m.ActiveColIndex--
				m.adjustScrollOffset()
			}
		case "j", "down":
			if len(m.Rows) > 0 && m.ActiveRowIndex < len(m.Rows)-1 {
				m.ActiveRowIndex++
			}
		case "k", "up":
			if m.ActiveRowIndex > 0 {
				m.ActiveRowIndex--
			}
		case "pgdown", "ctrl+d":
			totalPages := (m.TotalRows + m.PageSize - 1) / m.PageSize
			if totalPages == 0 {
				totalPages = 1
			}
			if m.CurrentPage < totalPages {
				m.CurrentPage++
				m.ActiveRowIndex = 0
				return m, func() tea.Msg {
					return PageChangedMsg{Page: m.CurrentPage}
				}
			}
		case "pgup", "ctrl+u":
			if m.CurrentPage > 1 {
				m.CurrentPage--
				m.ActiveRowIndex = 0
				return m, func() tea.Msg {
					return PageChangedMsg{Page: m.CurrentPage}
				}
			}
		case "d":
			if len(m.Rows) > 0 {
				m.ConfirmDelete = true
			}
		case "n":
			return m, func() tea.Msg {
				return CreateRowMsg{}
			}
		case "enter":
			if len(m.Rows) > 0 && !m.SchemaMode {
				m.EditMode = true
				initialVal := ""
				if m.ActiveRowIndex < len(m.Rows) && m.ActiveColIndex < len(m.Headers) {
					initialVal = m.Rows[m.ActiveRowIndex][m.ActiveColIndex]
				}
				m.EditInput.SetValue(initialVal)
				m.EditInput.Focus()
				return m, textinput.Blink
			}
		case "c":
			if len(m.Headers) > 0 {
				csvStr := export.ToCSV(m.Headers, m.Rows)
				err := clipboard.WriteAll(csvStr)
				msgText := "Table copied to clipboard as CSV!"
				if err != nil {
					msgText = "Error copying: " + err.Error()
				}
				return m, func() tea.Msg {
					return StatusMsg{Message: msgText}
				}
			}
		}
	}
	return m, nil
}

func truncateString(s string, l int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) > l {
		if l > 3 {
			return s[:l-3] + "..."
		}
		return s[:l]
	}
	return s
}

func (m GridModel) renderSchemaView() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Schema Table: %s\n\n", m.ActiveTable))

	// Render DDL block
	if m.SchemaDDL != "" {
		s.WriteString("CREATE STATEMENT:\n")
		s.WriteString(m.SchemaDDL + "\n\n")
	}

	s.WriteString("COLUMNS:\n")
	// Columns headers
	s.WriteString("+-----------------+-----------------+----------+----------+-----+\n")
	s.WriteString("| Column Name     | Type            | Not Null | Default  | PK  |\n")
	s.WriteString("+-----------------+-----------------+----------+----------+-----+\n")

	for _, col := range m.SchemaCols {
		pkStr := ""
		if col.IsPK {
			pkStr = "Yes"
		}
		nnStr := "No"
		if col.NotNull {
			nnStr = "Yes"
		}
		s.WriteString(fmt.Sprintf("| %-15s | %-15s | %-8s | %-8s | %-3s |\n",
			truncateString(col.Name, 15),
			truncateString(col.Type, 15),
			nnStr,
			truncateString(col.DefaultVal, 8),
			pkStr,
		))
	}
	s.WriteString("+-----------------+-----------------+----------+----------+-----+\n")

	return s.String()
}

func (m GridModel) View() string {
	if m.SchemaMode {
		return m.renderSchemaView()
	}

	if len(m.Headers) == 0 {
		return "No data loaded or table is empty."
	}

	var s strings.Builder

	// Calculate visible columns
	visibleCols := []int{}
	currentWidth := 0
	for i := m.ScrollOffset; i < len(m.Headers); i++ {
		colW := m.ColWidths[i] + 2 // padding
		if m.Width > 0 && currentWidth+colW+1 > m.Width && len(visibleCols) > 0 {
			break
		}
		visibleCols = append(visibleCols, i)
		currentWidth += colW + 1
	}

	if len(visibleCols) == 0 && len(m.Headers) > 0 {
		visibleCols = append(visibleCols, m.ScrollOffset)
	}

	// Render Top Border
	s.WriteString("+")
	for _, idx := range visibleCols {
		s.WriteString(strings.Repeat("-", m.ColWidths[idx]+2) + "+")
	}
	s.WriteString("\n|")
	for _, idx := range visibleCols {
		s.WriteString(fmt.Sprintf(" %-*s |", m.ColWidths[idx], m.Headers[idx]))
	}
	s.WriteString("\n+")
	for _, idx := range visibleCols {
		s.WriteString(strings.Repeat("-", m.ColWidths[idx]+2) + "+")
	}
	s.WriteString("\n")

	// Render Rows
	for rIdx, row := range m.Rows {
		s.WriteString("|")
		for _, idx := range visibleCols {
			val := ""
			if idx < len(row) {
				val = row[idx]
			}
			cellVal := truncateString(val, m.ColWidths[idx])

			var cellStr string
			if rIdx == m.ActiveRowIndex && idx == m.ActiveColIndex && m.Focused {
				cellStr = fmt.Sprintf("[%s]", cellVal)
				extra := (m.ColWidths[idx] + 2) - len(cellStr)
				if extra > 0 {
					cellStr = cellStr + strings.Repeat(" ", extra)
				}
			} else {
				cellStr = fmt.Sprintf(" %-*s ", m.ColWidths[idx], cellVal)
			}
			s.WriteString(cellStr + "|")
		}
		s.WriteString("\n")
	}

	// Bottom Border
	s.WriteString("+")
	for _, idx := range visibleCols {
		s.WriteString(strings.Repeat("-", m.ColWidths[idx]+2) + "+")
	}
	s.WriteString("\n")

	if m.ConfirmDelete {
		s.WriteString(" ⚠️  Delete row? Press 'y' to confirm, any other key to cancel.\n")
	} else if m.EditMode {
		s.WriteString(fmt.Sprintf(" Edit Cell: %s (Enter to confirm, Esc to cancel)\n", m.EditInput.View()))
	}

	// Pagination info bar
	totalPages := (m.TotalRows + m.PageSize - 1) / m.PageSize
	if totalPages == 0 {
		totalPages = 1
	}
	s.WriteString(fmt.Sprintf(" Page %d/%d | Total Rows: %d", m.CurrentPage, totalPages, m.TotalRows))
	if m.ScrollOffset > 0 || len(visibleCols) < len(m.Headers) {
		s.WriteString(fmt.Sprintf(" (Cols: %d-%d of %d)", m.ScrollOffset+1, m.ScrollOffset+len(visibleCols), len(m.Headers)))
	}
	s.WriteString("\n")

	return s.String()
}
