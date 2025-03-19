package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	model tea.Model
}

func RootScreen() RootModel {
	var rootModel tea.Model
	configScreen := ConfigurationScreen()
	rootModel = &configScreen
	return (RootModel{model: rootModel})
}
func (m RootModel) Init() tea.Cmd {
	return m.model.Init()
}
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.model.Update(msg)
}
func (m RootModel) View() string {
	return m.model.View()
}
func (m RootModel) SwitchScreen(model tea.Model) (tea.Model, tea.Cmd) {
	m.model = model
	return m.model, m.model.Init()
}

func main() {
	p := tea.NewProgram(RootScreen(), tea.WithMouseCellMotion(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oh dear, something bad happened and it crashed: %v", err)
		os.Exit(1)

	}
}
