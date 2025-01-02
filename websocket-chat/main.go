package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// Define a WebSocket upgrader
// spesifik ip banlama burada olabilir

var bannedIPs = map[string]bool{
	"192.168.1.100": true, // Example banned IP
	"10.0.0.5":      true, // Another banned IP
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Get the client's IP address
		ip := r.RemoteAddr

		fmt.Printf(ip)

		// Check if the IP is banned
		if bannedIPs[ip] {
			log.Printf("Banned IP tried to connect: %s\n", ip)
			return false // Reject the connection
		}

		return true // Allow the connection
	},
}

// Define a client struct to hold WebSocket connections
// burada user bilgileri tanımlanabilir
type Client struct {
	conn *websocket.Conn
	send chan []byte
	ip   string
}

// Define a chat server to manage clients
// burada chat room bilgileri tanımlanaiblir.
type ChatServer struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// Create a new chat server
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run the chat server
// burada kim ne zaman katılmış ne zaman odadan çıkmış loglar tutulabilir
func (cs *ChatServer) Run() {
	for {
		select {
		case client := <-cs.register:
			fmt.Printf("odaya girdi\n")
			fmt.Print(client)
			cs.clients[client] = true
		case client := <-cs.unregister:
			fmt.Printf("odadan çıktı\n")
			if _, ok := cs.clients[client]; ok {
				delete(cs.clients, client)
				close(client.send)
			}
		case message := <-cs.broadcast:
			for client := range cs.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(cs.clients, client)
				}
			}
		}
	}
}

// Handle WebSocket connections
func (cs *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("Error extracting IP:", err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256), ip: ip}
	cs.register <- client

	go client.WritePump(cs)
	go client.ReadPump(cs)

	client.send <- []byte("Your IP: " + ip)
}

// Read messages from the client
func (c *Client) ReadPump(cs *ChatServer) {

	// kullanıcı ayrılınca bu çalışıyor
	defer func() {
		fmt.Print("odadan ayrıldı tekrar")
		fmt.Print(cs.unregister)
		cs.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		formattedMessage := fmt.Sprintf("[%s]: %s", c.ip, string(message))
		cs.broadcast <- []byte(formattedMessage)
	}
}

// Write messages to the client
func (c *Client) WritePump(cs *ChatServer) {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}
}

func main() {
	chatServer := NewChatServer()
	go chatServer.Run()

	// Serve the WebSocket endpoint
	http.HandleFunc("/ws", chatServer.HandleWebSocket)

	// Serve a simple HTML page for testing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Start the HTTP server
	log.Println("Chat server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
