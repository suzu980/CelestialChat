package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type ConfigModel struct {
	form         *huh.Form
	ip           string
	port         string
	display_name string
}

func ConfigurationScreen() ConfigModel {
	return ConfigModel{
		form: huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("IP").Key("ip"), huh.NewInput().Title("Port").Key("port"), huh.NewInput().Title("Name").Key("display_name")),
		),
	}
}
func (m ConfigModel) Init() tea.Cmd {
	return m.form.Init()
}
func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	form, cmd := m.form.Update((msg))
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
		ip := m.form.GetString("ip")
		port := m.form.GetString("port")
		display_name := m.form.GetString("display_name")
		chatScreen := ChatScreen(ip, port, display_name)
		return RootScreen().SwitchScreen(&chatScreen)

	}
	return m, cmd
}

func (m ConfigModel) View() string {
	if m.form.State == huh.StateCompleted {
		ip := m.form.GetString("ip")
		port := m.form.GetString("port")
		display_name := m.form.GetString("display_name")
		return fmt.Sprintf("Hello %s! You entered: %s:%s\nPress any key to exit.", display_name, ip, port)
	}
	return m.form.View()
}
