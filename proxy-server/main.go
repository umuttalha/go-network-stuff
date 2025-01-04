package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

// curl -x http://localhost:8080 "https://karararama.yargitay.gov.tr/getDokuman?id=61850300" -v

func handleProxy(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Received %s request for: %s", r.Method, r.URL)

	if r.Method == http.MethodConnect {
		handleHTTPSTunnel(w, r)
		return
	}

	// Handle regular HTTP requests
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Printf("[ERROR] Failed to forward request: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleHTTPSTunnel(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Setting up HTTPS tunnel to: %s", r.Host)

	// Step 1: Connect to the target server
	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to target server %s: %v", r.Host, err)
		http.Error(w, fmt.Sprintf("Failed to connect to %s", r.Host), http.StatusServiceUnavailable)
		return
	}
	defer targetConn.Close()
	log.Printf("[DEBUG] Successfully connected to target: %s", r.Host)

	// Step 2: Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("[ERROR] ResponseWriter doesn't support hijacking")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("[ERROR] Failed to hijack connection: %v", err)
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()
	log.Printf("[DEBUG] Successfully hijacked client connection")

	// Step 3: Send 200 Connection Established
	response := []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	_, err = clientConn.Write(response)
	if err != nil {
		log.Printf("[ERROR] Failed to send 200 response: %v", err)
		return
	}
	log.Printf("[DEBUG] Sent 200 Connection Established response")

	// Step 4: Start bidirectional copy
	done := make(chan bool, 2)

	// Client to target
	go func() {
		io.Copy(targetConn, clientConn)
		done <- true
	}()

	// Target to client
	go func() {
		io.Copy(clientConn, targetConn)
		done <- true
	}()

	// Wait for either direction to finish
	<-done
	log.Printf("[DEBUG] Tunnel closed for %s", r.Host)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleProxy),
	}

	log.Printf("Starting proxy server on :8080")
	log.Fatal(server.ListenAndServe())
}
