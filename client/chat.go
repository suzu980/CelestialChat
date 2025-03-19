package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

type keyMap struct {
	Quit key.Binding
}
type websocketMsg struct {
	content string
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

	conn *websocket.Conn

	err error
}

func listenForWSMessages(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return tea.Quit
		}
		return websocketMsg{content: string(message)}
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit}}
}

func ChatScreen(
	ip string, port string, display_name string, width int, height int,
) (ChatModel, error) {
	var keys = keyMap{
		Quit: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Quit the program")),
	}
	var ip_, port_ string
	if ip == "" {
		ip_ = "0.0.0.0"
	} else {
		ip_ = ip
	}
	if port == "" {
		port_ = "6969"
	} else {
		port_ = port
	}
	serverURL := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%s", ip_, port_), Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		return ChatModel{err: err}, err
	}
	// Send name to server

	if err := conn.WriteMessage(websocket.TextMessage, []byte(display_name)); err != nil {
		fmt.Println("Failed to send username", err)
		return ChatModel{err: err}, err
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

	// Welcome message
	welcome := fmt.Sprintf("Welcome to %s", lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("✨ Celestial Chat ✨"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("Hope you enjoy your stay here <3")

	h := help.New()
	init_message_log := []string{"", welcome, "", info, "", ""}

	vp.Width = width
	vp.Height = height - textarea.Height() - lipgloss.Height(gap) - lipgloss.Height(h.View(keys))
	textarea.SetWidth(width)
	if len(init_message_log) > 0 {
		// Wrap content before setting it.
		vp.SetContent(lipgloss.NewStyle().Width(vp.Width).Render(strings.Join(init_message_log, "\n")))
	}
	vp.GotoBottom()
	chat_model := ChatModel{
		ip:           ip,
		port:         port,
		display_name: display_name,
		keys:         keys,
		help:         help.New(),
		chat_area:    textarea,
		viewport:     vp,
		sender_style: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		message_log:  init_message_log,
		conn:         conn,
	}
	return chat_model, nil

}

const gap = "\n\n"

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, listenForWSMessages(m.conn))
}
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			default:
				if m.conn != nil {
					m.conn.Close()
				}
				return m, tea.Quit
			}
		}
	}
	var (
		helpCmd tea.Cmd
		chatCmd tea.Cmd
		vpCmd   tea.Cmd
	)
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.message_log, "\n")))
	//Handle window resize
	switch msg := msg.(type) {
	case websocketMsg:
		m.message_log = append(m.message_log, msg.content)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.message_log, "\n")))
		// Restart the websocket
		cmds = append(cmds, listenForWSMessages(m.conn))
		m.viewport.GotoBottom()
	case tea.WindowSizeMsg:
		currentWidth := msg.Width
		currentHeight := msg.Height - m.chat_area.Height() - lipgloss.Height(gap) - lipgloss.Height(m.help.View(m.keys))
		if currentWidth <= 0 {
			m.viewport.Width = 3
		} else {
			m.viewport.Width = currentWidth
		}

		if currentHeight <= 0 {
			m.viewport.Height = 3
		} else {
			m.viewport.Height = currentHeight
		}
		m.chat_area.SetWidth(msg.Width)
		if len(m.message_log) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.message_log, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			written_message := m.sender_style.Render(fmt.Sprintf("%s: ", m.display_name)) + m.chat_area.Value()
			if err := m.conn.WriteMessage(websocket.TextMessage, []byte(written_message)); err != nil {
				m.err = err
			}
			m.chat_area.Reset()
			m.viewport.GotoBottom()
		}
	}
	m.help, helpCmd = m.help.Update(msg)
	m.chat_area, chatCmd = m.chat_area.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, helpCmd, chatCmd, vpCmd)
	return m, tea.Batch(cmds...)
}
func (m ChatModel) View() string {
	if m.err != nil {
		if m.conn != nil {
			m.conn.Close()
		}
		return fmt.Sprintf("\nAn error has occured:\n%s\nPress any key to exit.", m.err)
	}
	chatlog_view := m.viewport.View()
	help_view := m.help.View(m.keys)
	chat_area := m.chat_area.View()
	return fmt.Sprintf("%s%s\n%s\n%s", chatlog_view, gap, chat_area, help_view)
}
