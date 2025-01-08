package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/natefinch/lumberjack"
)

// Logger setup
func setupLogger() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/foo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 0,   // infinite backups
		MaxAge:     0,
		Compress:   true, // disabled by default
	})
}

// real ip from user cloudflare header
func getRealIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	return r.RemoteAddr
}

// Middleware for logging requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getRealIP(r)

		// Log POST request body
		if r.Method == http.MethodPost {
			// post request için sırası tarih, Method, Endpoint,Query,Ip
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
			} else {

				log.Printf(",%s ,%s ,%s ,%s\n", r.Method, r.URL.Path, string(bodyBytes), clientIP)

				// Replace the original body so the handler can read it (very importatn)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		} else {
			// get request için sırası tarih, Method, Endpoint,Query,Ip
			log.Printf(",%s ,%s ,%s ,%s\n", r.Method, r.URL.Path, r.URL.RawQuery, clientIP)

		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the home page!"))
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Write([]byte("Data received successfully!"))
	} else {
		w.Write([]byte("This endpoint only supports POST requests."))
	}
}

func main() {
	// Logger setup
	setupLogger()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register endpoints
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/submit", submitHandler) // For POST requests

	// Wrap the mux with the logging middleware
	loggedMux := loggingMiddleware(mux)

	// Start the server
	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
