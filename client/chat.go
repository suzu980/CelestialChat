package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Quit key.Binding
}

type ChatModel struct {
	ip           string
	port         string
	display_name string

	keys      keyMap
	help      help.Model
	chat_area textarea.Model
	viewport  viewport.Model

	sender_style lipgloss.Style
	message_log  []string
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit}}
}

func ChatScreen(ip string, port string, display_name string) ChatModel {
	var keys = keyMap{
		Quit: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Quit the program")),
	}
	textarea := textarea.New()
	textarea.Placeholder = fmt.Sprintf("Chatting as %s", display_name)
	textarea.Focus()
	textarea.Prompt = "┃ "
	textarea.CharLimit = 280
	textarea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	textarea.ShowLineNumbers = false
	textarea.SetWidth(30)
	textarea.SetHeight(3)
	textarea.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)
	vp.KeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"), // Use "j" instead of "pgdown"
			key.WithHelp("j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"), // Use "k" instead of "pgup"
			key.WithHelp("k", "scroll up"),
		),
	}
	vp.MouseWheelEnabled = true

	return ChatModel{
		ip:           ip,
		port:         port,
		display_name: display_name,
		keys:         keys,
		help:         help.New(),
		chat_area:    textarea,
		viewport:     vp,
		sender_style: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		message_log:  []string{},
	}
}

const gap = "\n\n"

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		helpCmd tea.Cmd
		chatCmd tea.Cmd
		vpCmd   tea.Cmd
	)
	m.help, helpCmd = m.help.Update(msg)
	m.chat_area, chatCmd = m.chat_area.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	//Handle window resize
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.chat_area.Height() - lipgloss.Height(gap) - lipgloss.Height(m.help.View(m.keys))
		m.chat_area.SetWidth(msg.Width)
		if len(m.message_log) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.message_log, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.message_log = append(m.message_log, m.sender_style.Render(fmt.Sprintf("%s: ", m.display_name))+m.chat_area.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.message_log, "\n")))
			m.chat_area.Reset()
			m.viewport.GotoBottom()
		}
	}
	return m, tea.Batch(vpCmd, helpCmd, chatCmd)
}
func (m ChatModel) View() string {
	// ip := m.ip
	// port := m.port
	// display_name := m.display_name
	chatlog_view := m.viewport.View()
	help_view := m.help.View(m.keys)
	chat_area := m.chat_area.View()
	// return fmt.Sprintf("Hello %s! Attempting to connect to server: ws://%s:%s/ws\n", display_name, ip, port)
	return fmt.Sprintf("%s%s\n%s\n%s", chatlog_view, gap, chat_area, help_view)
	// return fmt.Sprintf("Hello %s! Attempting to connect to server: ws://[REDACTED]:%s/ws\n%s\n%s", display_name, port, chat_area, help_view)
}
