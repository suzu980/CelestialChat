package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	// Ask for a username
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	// Send username to server
	if err := conn.WriteMessage(websocket.TextMessage, []byte(username)); err != nil {
		log.Fatal("Failed to send username:", err)
	}

	fmt.Println("Connected to chat as:", username)
	fmt.Println("Type messages and press Enter to send.")

	// Goroutine to receive messages
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			fmt.Printf("\r%s\n> ", string(msg)) // Clear prompt and print received message
		}
	}()

	// Read and send user messages
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			msg := scanner.Text()

			// Move cursor up, clear line to prevent duplicate prompts
			fmt.Print("\033[1A\033[K")

			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}
