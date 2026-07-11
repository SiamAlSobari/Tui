package components

import (
	"fmt"
	"strings"
	"tui-sqlite/internal/db"
	"tui-sqlite/internal/export"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type PageChangedMsg struct {
	Page int
}

type StatusMsg struct {
	Message string
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
}

func NewGrid() GridModel {
	return GridModel{
		Headers:      []string{},
		Rows:         [][]string{},
		PageSize:     50,
		CurrentPage:  1,
		ColWidths:    make(map[int]int),
		ScrollOffset: 0,
		SchemaMode:   false,
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

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l", "right":
			if m.ScrollOffset < len(m.Headers)-1 {
				m.ScrollOffset++
			}
		case "h", "left":
			if m.ScrollOffset > 0 {
				m.ScrollOffset--
			}
		case "pgdown", "ctrl+d":
			totalPages := (m.TotalRows + m.PageSize - 1) / m.PageSize
			if totalPages == 0 {
				totalPages = 1
			}
			if m.CurrentPage < totalPages {
				m.CurrentPage++
				return m, func() tea.Msg {
					return PageChangedMsg{Page: m.CurrentPage}
				}
			}
		case "pgup", "ctrl+u":
			if m.CurrentPage > 1 {
				m.CurrentPage--
				return m, func() tea.Msg {
					return PageChangedMsg{Page: m.CurrentPage}
				}
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
	for _, row := range m.Rows {
		s.WriteString("|")
		for _, idx := range visibleCols {
			val := ""
			if idx < len(row) {
				val = row[idx]
			}
			s.WriteString(fmt.Sprintf(" %-*s |", m.ColWidths[idx], truncateString(val, m.ColWidths[idx])))
		}
		s.WriteString("\n")
	}

	// Bottom Border
	s.WriteString("+")
	for _, idx := range visibleCols {
		s.WriteString(strings.Repeat("-", m.ColWidths[idx]+2) + "+")
	}
	s.WriteString("\n")

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
