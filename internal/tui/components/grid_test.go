package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGridAutoColumnWidth(t *testing.T) {
	g := NewGrid()
	headers := []string{"id", "username", "email"}
	rows := [][]string{
		{"1", "alice", "alice@example.com"},
		{"2", "bob", "bob@example.com"},
		{"3000", "charlie_long_name", "charlie@example.com"},
	}

	g.SetData(headers, rows, 3)

	// Verify column widths (should fit the maximum length in each column)
	// Column 0: "id" (2), "1" (1), "2" (1), "3000" (4) => Max 4. With padding, should be at least 4.
	// Column 1: "username" (8), "alice" (5), "bob" (3), "charlie_long_name" (17) => Max 17.
	// Column 2: "email" (5), "alice@example.com" (17), "bob@example.com" (15), "charlie@example.com" (19) => Max 19.
	
	if g.ColWidths[0] < 4 {
		t.Errorf("expected col 0 width >= 4, got %d", g.ColWidths[0])
	}
	if g.ColWidths[1] < 17 {
		t.Errorf("expected col 1 width >= 17, got %d", g.ColWidths[1])
	}
	if g.ColWidths[2] < 19 {
		t.Errorf("expected col 2 width >= 19, got %d", g.ColWidths[2])
	}
}

func TestGridHorizontalScrolling(t *testing.T) {
	g := NewGrid()
	headers := []string{"col1", "col2", "col3", "col4"}
	rows := [][]string{
		{"a", "b", "c", "d"},
	}
	g.SetData(headers, rows, 1)
	g.Width = 30 // Set a small container width

	// Initially horizontal offset is 0
	if g.ScrollOffset != 0 {
		t.Errorf("expected initial ScrollOffset to be 0, got %d", g.ScrollOffset)
	}

	// Press 'l' or Right arrow to scroll right
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if g.ScrollOffset != 1 {
		t.Errorf("expected ScrollOffset to be 1 after scrolling right, got %d", g.ScrollOffset)
	}

	// Press 'h' or Left arrow to scroll left
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if g.ScrollOffset != 0 {
		t.Errorf("expected ScrollOffset to return to 0 after scrolling left, got %d", g.ScrollOffset)
	}
}

func TestGridPagination(t *testing.T) {
	g := NewGrid()
	g.PageSize = 2
	g.TotalRows = 5 // Total 5 rows, so 3 pages (2, 2, 1)

	if g.CurrentPage != 1 {
		t.Errorf("expected initial page to be 1, got %d", g.CurrentPage)
	}

	// Go to next page (PgDn / ctrl+d)
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if g.CurrentPage != 2 {
		t.Errorf("expected page 2 after PgDn, got %d", g.CurrentPage)
	}

	// Go to next page again
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if g.CurrentPage != 3 {
		t.Errorf("expected page 3 after second PgDn, got %d", g.CurrentPage)
	}

	// Go to next page again (should clamp to max page 3)
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if g.CurrentPage != 3 {
		t.Errorf("expected page to clamp to 3, got %d", g.CurrentPage)
	}

	// Go back (PgUp / ctrl+u)
	g, _ = g.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if g.CurrentPage != 2 {
		t.Errorf("expected page 2 after PgUp, got %d", g.CurrentPage)
	}
}
