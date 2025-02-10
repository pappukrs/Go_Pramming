package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (for development)
	},
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan Message)           // Broadcast channel

// Define message structure
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	http.HandleFunc("/", handleConnections)

	go handleMessages()

	fmt.Println("WebSocket server started on :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()

	// Register the new client
	clients[ws] = true

	for {
		var msg Message
		// Read a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			delete(clients, ws)
			break
		}
		// Send the received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Get the next message from the broadcast channel
		msg := <-broadcast
		// Send the message to all connected clients
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
