package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type command struct {
	sender *websocket.Conn
	name   string
	args   string
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func Strip(str string) string {
	return re.ReplaceAllString(str, "")
}

// Upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Map to store clients and their usernames
var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan string)
var commands = make(chan command)
var mu sync.Mutex // Mutex to avoid concurrent map writes

// Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Read the first message as the username
	_, usernameBytes, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read username:", err)
		return
	}
	username := string(usernameBytes)

	// Add client to the map
	mu.Lock()
	clients[conn] = username
	userCount := len(clients)
	mu.Unlock()

	// Broadcast join message
	log.Printf("%s has joined the chat!\n", username)
	broadcast <- fmt.Sprintf("\033[33m\nA wild %s has joined the chat! (%d online)\n\033[0m", username, userCount)

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s disconnected.\n", username)
			break
		}
		formattedMessage := Strip(string(msg))
		_, after, found := strings.Cut(formattedMessage, ": ")
		if found {
			if strings.HasPrefix(after, "/") {
				// Parse and queue the command
				parts := strings.SplitN(after, " ", 2)
				cmdName := parts[0]
				cmdArgs := ""
				if len(parts) > 1 {
					cmdArgs = parts[1]
				}
				commands <- command{sender: conn, name: cmdName, args: cmdArgs}

			} else {
				// Broadcast message to all clients
				broadcast <- fmt.Sprintf("%s", string(msg))
			}

		} else {
			// Broadcast message to all clients
			broadcast <- fmt.Sprintf("%s", string(msg))
		}
	}

	// Handle disconnect
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	broadcast <- fmt.Sprintf("\033[31m\nOh dear, %s has disconnected the chat. (%d online)\n\033[0m", username, userCount)
}

// Handle broadcasting messages
func handleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func handleCommands() {
	for cmd := range commands {
		mu.Lock()
		username, exists := clients[cmd.sender]
		mu.Unlock()
		if !exists {
			continue
		}
		switch cmd.name {
		case "/list":
			// Send a list of online users
			var userList []string
			mu.Lock()
			for _, name := range clients {
				userList = append(userList, name)
			}
			mu.Unlock()
			response := fmt.Sprintf("\033[36m\nCurrent Online Users (%d): %s\n\033[0m", len(userList), strings.Join(userList, ", "))
			cmd.sender.WriteMessage(websocket.TextMessage, []byte(response))

		case "/me":
			// Emote message
			if cmd.args == "" {
				cmd.sender.WriteMessage(websocket.TextMessage, []byte("Usage: /me <message>"))
				continue
			}
			broadcast <- fmt.Sprintf("\033[35m\n* %s %s *\n\033[0m", username, cmd.args)
		case "/em":
			// Emote message
			if cmd.args == "" {
				cmd.sender.WriteMessage(websocket.TextMessage, []byte("Usage: /em <message>"))
				continue
			}
			broadcast <- fmt.Sprintf("\033[35m\n* %s *\n\033[0m", cmd.args)

		default:
			response := fmt.Sprintf("\nUnknown command: %s\n", cmd.name)
			cmd.sender.WriteMessage(websocket.TextMessage, []byte(response))
		}
	}
}

func main() {
	// Prompt for IP and Port
	var ip, port string
	fmt.Print("Enter the IP address to bind the server: ")
	fmt.Scanln(&ip)
	if ip == "" {
		ip = "0.0.0.0" // Default to all available network interfaces
	}

	fmt.Print("Enter the port to listen on: ")
	fmt.Scanln(&port)
	if port == "" {
		port = "6969" // Default port
	}
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	go handleCommands()

	serverAddr := fmt.Sprintf("%s:%s", ip, port)
	fmt.Printf("WebSocket server starting on ws://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
