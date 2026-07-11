package tui

import (
	"fmt"
	"dbbee/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing layout..."
	}

	// 1. Render Header
	dbName := "No Database"
	if m.DB != nil {
		dbName = m.DB.Path
	}
	headerText := fmt.Sprintf(" TuiSqlite v1.0.0 [ database: %s ]", dbName)
	header := styles.HeaderStyle.Render(headerText)

	// 2. Main Body height calculation
	// We reserve 1 line for header, 1 line for footer, and 2 lines for borders/margins.
	bodyHeight := m.Height - 4
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	// Determine sidebar width (fixed 25 columns)
	sidebarWidth := 25

	// Determine right content width
	rightWidth := m.Width - sidebarWidth - 6 // minus borders/padding
	if rightWidth < 10 {
		rightWidth = 10
	}

	// Apply styles based on active focus
	sidebarBorder := styles.SidebarStyle
	if m.ActiveTab == SidebarTab {
		sidebarBorder = styles.SidebarFocusedStyle
	}
	sidebarBorder = sidebarBorder.Height(bodyHeight).Width(sidebarWidth)

	rightPanelBorder := styles.MainAreaStyle
	if m.ActiveTab == GridTab || m.ActiveTab == EditorTab {
		rightPanelBorder = styles.MainAreaFocusedStyle
	}
	rightPanelBorder = rightPanelBorder.Height(bodyHeight).Width(rightWidth)

	// Render Sidebar View
	sidebarView := sidebarBorder.Render(m.Sidebar.View())

	// Set grid boundaries so auto-column calculation and pagination fit correctly
	m.Grid.Width = rightWidth
	m.Grid.Height = bodyHeight

	// Render Right Panel Content
	var rightViewContent string
	if m.ActiveTab == EditorTab {
		rightViewContent = m.Editor.View()
	} else {
		rightViewContent = m.Grid.View()
	}
	rightView := rightPanelBorder.Render(rightViewContent)

	// Join horizontally
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, rightView)

	// 3. Render Footer
	var footerText string
	switch m.ActiveTab {
	case SidebarTab:
		footerText = " [Tab] Focus Grid | [/] Filter | [j/k] Navigate | [q] Quit"
	case GridTab:
		footerText = " [Tab] Focus Editor | [c] Copy CSV | [h/l] Scroll Cols | [PgUp/PgDn] Paging | [q] Quit"
	case EditorTab:
		footerText = " [Tab] Focus Sidebar | [Ctrl+J/Ctrl+Enter] Run Query | [q] Quit"
	default:
		footerText = " [Tab] Switch Panels | [q] Quit"
	}

	footerContent := footerText
	if m.StatusMessage != "" {
		spinnerPrefix := ""
		if m.Loading {
			spinnerPrefix = m.Spinner.View() + " "
		}
		footerContent = fmt.Sprintf(" Status: %s%s\n%s", spinnerPrefix, m.StatusMessage, footerText)
	} else if m.Loading {
		footerContent = fmt.Sprintf(" Status: %s Loading...\n%s", m.Spinner.View(), footerText)
	}
	footer := styles.FooterStyle.Render(footerContent)

	// Combine vertically
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
