package tui

import (
	"fmt"
	"strings"
	"testing"
	"tui-sqlite/internal/db"
	"tui-sqlite/internal/tui/components"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIFocusSwitching(t *testing.T) {
	m := NewModel(nil)
	if m.ActiveTab != SidebarTab {
		t.Errorf("expected initial tab to be SidebarTab, got %v", m.ActiveTab)
	}

	// 1. Tab should switch to GridTab
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(Model)
	if m.ActiveTab != GridTab {
		t.Errorf("expected Tab to switch to GridTab, got %v", m.ActiveTab)
	}

	// 2. Tab should switch to EditorTab
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(Model)
	if m.ActiveTab != EditorTab {
		t.Errorf("expected Tab to switch to EditorTab, got %v", m.ActiveTab)
	}

	// 3. Tab should switch back to SidebarTab
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(Model)
	if m.ActiveTab != SidebarTab {
		t.Errorf("expected Tab to loop back to SidebarTab, got %v", m.ActiveTab)
	}

	// 4. Shift+Tab should go backward: SidebarTab -> EditorTab
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = res.(Model)
	if m.ActiveTab != EditorTab {
		t.Errorf("expected Shift+Tab to switch to EditorTab, got %v", m.ActiveTab)
	}
}

func TestSidebarFiltering(t *testing.T) {
	tables := []db.TableInfo{
		{Name: "users", Type: "table", RowCount: 10},
		{Name: "posts", Type: "table", RowCount: 20},
		{Name: "comments", Type: "table", RowCount: 30},
	}

	sb := components.NewSidebar()
	sb.Tables = tables
	sb.Filtered = tables
	sb.ActiveIndex = 0

	// 1. Press '/' to activate filter
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if !sb.FilterActive {
		t.Error("expected FilterActive to be true after pressing '/'")
	}

	// 2. Press 'p'
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if sb.FilterInput.Value() != "p" {
		t.Errorf("expected filter value to be 'p', got %q", sb.FilterInput.Value())
	}
	if len(sb.Filtered) != 1 || sb.Filtered[0].Name != "posts" {
		t.Errorf("expected filtered list to contain only 'posts', got %v", sb.Filtered)
	}

	// 3. Press 'Esc' to close/reset filter
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if sb.FilterActive {
		t.Error("expected FilterActive to be false after Esc")
	}
	if sb.FilterInput.Value() != "" {
		t.Errorf("expected filter input to be cleared, got %q", sb.FilterInput.Value())
	}
	if len(sb.Filtered) != 3 {
		t.Errorf("expected filtered list to reset to all tables, got %v", sb.Filtered)
	}
}

func TestSidebarNavigation(t *testing.T) {
	tables := []db.TableInfo{
		{Name: "users", Type: "table", RowCount: 10},
		{Name: "posts", Type: "table", RowCount: 20},
	}

	sb := components.NewSidebar()
	sb.Tables = tables
	sb.Filtered = tables
	sb.ActiveIndex = 0

	// 1. Press 'j' to navigate down
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sb.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex to be 1 after 'j', got %d", sb.ActiveIndex)
	}

	// 2. Press 'j' again (should not change active index because we're at the bottom)
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sb.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex to remain 1 at boundary, got %d", sb.ActiveIndex)
	}

	// 3. Press 'k' to navigate up
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if sb.ActiveIndex != 0 {
		t.Errorf("expected ActiveIndex to be 0 after 'k', got %d", sb.ActiveIndex)
	}

	// 4. Press Up Arrow (should also work)
	sb.ActiveIndex = 0
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sb.ActiveIndex != 0 {
		t.Errorf("expected ActiveIndex to remain 0 on Up arrow at top, got %d", sb.ActiveIndex)
	}

	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sb.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex to be 1 on Down arrow, got %d", sb.ActiveIndex)
	}
}

func TestTUIWindowSize(t *testing.T) {
	m := NewModel(nil)
	if m.Width != 0 || m.Height != 0 {
		t.Errorf("expected initial dimensions to be 0, got %dx%d", m.Width, m.Height)
	}

	// Sending a WindowSizeMsg
	res, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = res.(Model)

	if m.Width != 80 || m.Height != 24 {
		t.Errorf("expected dimensions to update to 80x24, got %dx%d", m.Width, m.Height)
	}

	// Verify view does not return initialization message anymore
	view := m.View()
	if view == "Initializing layout..." {
		t.Error("expected view to render layout, but got initialization message")
	}
}

func TestTUIQueryExecution(t *testing.T) {
	m := NewModel(nil)

	// Simulate RunQueryMsg
	res, _ := m.Update(components.RunQueryMsg{SQL: "SELECT 1"})
	m = res.(Model)
	if m.StatusMessage != "Executing query..." {
		t.Errorf("expected executing status, got %q", m.StatusMessage)
	}

	// Simulate RunQueryResultMsg success
	headers := []string{"num"}
	rows := [][]string{{"1"}}
	res, _ = m.Update(RunQueryResultMsg{Headers: headers, Rows: rows})
	m = res.(Model)

	if m.Grid.ActiveTable != "Custom Query" {
		t.Errorf("expected Grid table to be 'Custom Query', got %q", m.Grid.ActiveTable)
	}
	if len(m.Grid.Headers) != 1 || m.Grid.Headers[0] != "num" {
		t.Errorf("unexpected headers in grid: %v", m.Grid.Headers)
	}
	if len(m.Grid.Rows) != 1 || m.Grid.Rows[0][0] != "1" {
		t.Errorf("unexpected rows in grid: %v", m.Grid.Rows)
	}
	if !strings.Contains(m.StatusMessage, "successfully") {
		t.Errorf("expected success status message, got %q", m.StatusMessage)
	}
}

func TestTUIQueryExecutionError(t *testing.T) {
	m := NewModel(nil)

	// Simulate error result
	res, _ := m.Update(RunQueryResultMsg{Err: fmt.Errorf("syntax error")})
	m = res.(Model)

	if !strings.Contains(m.StatusMessage, "SQL Error: syntax error") {
		t.Errorf("expected error status message, got %q", m.StatusMessage)
	}
}
