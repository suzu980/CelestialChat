package main

//
// import (
// 	"bufio"
// 	"fmt"
// 	"log"
// 	"os"
//
// 	"github.com/gorilla/websocket"
// )
//
// func main() {
// 	// Prompt for IP and Port
// 	var ip, port string
// 	fmt.Print("Enter the IP address to bind the server: ")
// 	fmt.Scanln(&ip)
// 	if ip == "" {
// 		ip = "0.0.0.0" // Default to all available network interfaces
// 	}
//
// 	fmt.Print("Enter the port to listen on: ")
// 	fmt.Scanln(&port)
// 	if port == "" {
// 		port = "8080" // Default port
// 	}
// 	serverAddr := fmt.Sprintf("ws://%s:%s/ws", ip, port)
//
// 	// Connect to the WebSocket server
// 	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
// 	if err != nil {
// 		log.Fatal("Connection error:", err)
// 	}
// 	defer conn.Close()
//
// 	// Ask for a username
// 	fmt.Print("Enter your name: ")
// 	scanner := bufio.NewScanner(os.Stdin)
// 	scanner.Scan()
// 	username := scanner.Text()
//
// 	// Send username to server
// 	if err := conn.WriteMessage(websocket.TextMessage, []byte(username)); err != nil {
// 		log.Fatal("Failed to send username:", err)
// 	}
//
// 	fmt.Println("==========================================")
// 	fmt.Println()
// 	fmt.Println("✨ Celestial Chat")
// 	fmt.Println("Connecting to:", serverAddr)
// 	fmt.Println("Welcome,", username, "!")
// 	fmt.Println()
// 	fmt.Println("==========================================")
// 	fmt.Println()
// 	fmt.Println("Type messages and press Enter to send.")
//
// 	// Goroutine to receive messages
// 	go func() {
// 		for {
// 			_, msg, err := conn.ReadMessage()
// 			if err != nil {
// 				log.Println("Read error:", err)
// 				return
// 			}
// 			fmt.Printf("\r%s\n> ", string(msg)) // Clear prompt and print received message
// 		}
// 	}()
//
// 	// Read and send user messages
// 	for {
// 		fmt.Print("> ")
// 		if scanner.Scan() {
// 			msg := scanner.Text()
//
// 			// Move cursor up, clear line to prevent duplicate prompts
// 			fmt.Print("\033[1A\033[K")
//
// 			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
// 				log.Println("Write error:", err)
// 				return
// 			}
// 		}
// 	}
// }
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
	p := tea.NewProgram(RootScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oh dear, something bad happened and it crashed: %v", err)
		os.Exit(1)

	}
}
