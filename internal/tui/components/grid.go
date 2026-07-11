package components

import tea "github.com/charmbracelet/bubbletea"

type GridModel struct{}

func NewGrid() GridModel {
	return GridModel{}
}

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	return m, nil
}

func (m GridModel) View() string {
	return ""
}
