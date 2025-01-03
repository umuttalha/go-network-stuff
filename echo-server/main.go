package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// for create basic server we can use nc -v localhost 8080 for create test server. this can be any http server too

// before using this code we can basically use "nc -v -l localhost 8080" command create listener instead of this
// but in this example we change to reader.ReadString('.') so detact the port until . charecter

func main() {
	// Step 1: Start listening on a specific port
	listener, err := net.Listen("tcp", ":8080") // ":8080" means listening on port 8080 on all interfaces
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		// Step 2: Accept a connection from a client
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New client connected!")

		// Step 3: Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Ensure the connection is closed when the function ends

	// Step 4: Read data from the client
	reader := bufio.NewReader(conn)
	for {
		// Read client message until there is .
		// normally using \n for until newline but i want this
		message, err := reader.ReadString('.')
		if err != nil {
			fmt.Println("Client disconnected or error:", err)
			return
		}

		// for more stupic example

		// if strings.Contains(message, "umut") {
		// 	fmt.Printf("Received message with 'umut': %s", message)
		// 	_, _ = conn.Write([]byte("Message contains 'umut'!\n"))
		// } else {
		// 	fmt.Printf("Received: %s", message)
		// 	_, _ = conn.Write([]byte("No 'umut' found in your message.\n"))
		// }

		// in there in sentence if contain umut or not we detect

		fmt.Printf("Received: %s", message)

		// Step 5: Echo the message back to the client
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message back:", err)
			return
		}
	}
}
