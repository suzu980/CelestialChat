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
	width        int
	height       int
}

func ConfigurationScreen() ConfigModel {
	return ConfigModel{
		form: huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("IP address to connect to: (Defaults to: localhost)").Key("ip"), huh.NewInput().Title("Port: (Defaults to: 6969 *nice*)").Key("port"), huh.NewInput().Title("Your display Name").Key("display_name")),
		),
		width:  30,
		height: 3,
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
	case tea.WindowSizeMsg:
		currentWidth := msg.Width
		currentHeight := msg.Height
		if currentWidth <= 0 {
			m.width = 3
		} else {
			m.width = currentWidth
		}

		if currentHeight <= 0 {
			m.height = 3
		} else {
			m.height = currentHeight
		}
	}
	form, cmd := m.form.Update((msg))
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		ip := m.form.GetString("ip")
		port := m.form.GetString("port")
		display_name := m.form.GetString("display_name")
		chatScreen, err := ChatScreen(ip, port, display_name, m.width, m.height)
		if err != nil {

			if _, ok := msg.(tea.KeyMsg); ok {
				return m, tea.Quit
			}
			return m, nil
		}
		return RootScreen().SwitchScreen(&chatScreen)

	}
	return m, cmd
}

func (m ConfigModel) View() string {
	if m.form.State == huh.StateCompleted {
		return fmt.Sprintf("If you're seeing this screen, it means that connection doesn't exist or something weird error happened.\nWhat happened? I may not know ¯\\_(ツ)_/¯\n\nPress any key to exit.")
	}
	return m.form.View()
}
