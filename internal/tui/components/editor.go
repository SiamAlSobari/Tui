package components

import tea "github.com/charmbracelet/bubbletea"

type EditorModel struct{}

func NewEditor() EditorModel {
	return EditorModel{}
}

func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	return m, nil
}

func (m EditorModel) View() string {
	return ""
}
