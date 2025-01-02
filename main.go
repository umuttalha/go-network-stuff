package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close the connection when the function returns

	fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	for {
		// Read data from the connection
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection from %s closed: %s\n", conn.RemoteAddr().String(), err)
			return
		}

		// Print the received message
		fmt.Printf("Received from %s: %s", conn.RemoteAddr().String(), message)

		// Echo the message back to the client
		conn.Write([]byte(message))
	}
}

func main() {
	// Define the address to listen on
	address := "127.0.0.1:8080"

	// Start listening on the specified address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("TCP Echo Server is listening on %s\n", address)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %s\n", err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}
