package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func main() {
	// Add file server for the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	
	// Handle WebSocket connections at /ws endpoint instead of root
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8081")
	err := http.ListenAndServe("192.168.1.7:8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		// Read message from client (SDP or ICE candidate)
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			delete(clients, ws)
			break
		}
		// Broadcast the message to all clients
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println("Error writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
