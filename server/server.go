package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Map to store clients and their usernames
var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan string)
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
	mu.Unlock()

	// Broadcast join message
	log.Printf("\n📢 %s has joined the chat!", username)
	broadcast <- fmt.Sprintf("\033[33m\nA wild %s has joined the chat!\n\033[0m", username)

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		log.Printf("'%s' message read.\n", msg)
		if err != nil {
			log.Printf("%s disconnected.\n", username)
			break
		}
		// Broadcast message to all clients
		broadcast <- fmt.Sprintf("%s", string(msg))
	}

	// Handle disconnect
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	broadcast <- fmt.Sprintf("\033[31m\nOh dear, %s has disconnected the chat.\n\033[0m", username)
}

// Handle broadcasting messages
func handleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
		log.Printf("'%s' sent.\n", msg)
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

	serverAddr := fmt.Sprintf("%s:%s", ip, port)
	fmt.Printf("WebSocket server starting on ws://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
