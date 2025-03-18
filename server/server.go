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
	broadcast <- fmt.Sprintf("📢 %s has joined the chat!", username)

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s disconnected.\n", username)
			break
		}
		// Broadcast message to all clients
		broadcast <- fmt.Sprintf("%s: %s", username, string(msg))
	}

	// Handle disconnect
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	broadcast <- fmt.Sprintf("❌ %s has left the chat.", username)
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

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	fmt.Println("WebSocket server running on ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
