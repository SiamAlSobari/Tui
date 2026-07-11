package styles

import "github.com/charmbracelet/lipgloss"

// Color constants
const (
	ColorPrimary   = lipgloss.Color("99")  // Violet/Purple
	ColorBgSidebar = lipgloss.Color("235") // Dark gray
	ColorSelection = lipgloss.Color("99")  // Highlight color
	ColorText      = lipgloss.Color("255") // White
	ColorMuted     = lipgloss.Color("243") // Muted gray
)

// Styling borders and boxes
var (
	SidebarStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)

	SidebarFocusedStyle = SidebarStyle.Copy().
		BorderForeground(ColorPrimary)

	MainAreaStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)

	MainAreaFocusedStyle = MainAreaStyle.Copy().
		BorderForeground(ColorPrimary)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(0, 1)
)
