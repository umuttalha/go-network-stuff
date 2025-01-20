package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket bağlantısını yükseltmek için bir upgrade fonksiyonu
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		// Sadece "https://example.com" ve "http://localhost:3000" origin'lerine izin ver

		// 1. Sadece HTTPS bağlantılarına izin ver
		// if r.TLS == nil {
		//     return false
		// }

		allowedOrigins := []string{
			"https://sisatma.com",
			"http://localhost:8080",
		}

		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			}
		}

		return false
	},
}

func main() {
	// WebSocket bağlantısını dinleyecek endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// HTTP bağlantısını WebSocket'e yükselt
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket bağlantısı kurulamadı:", err)
			return
		}
		defer conn.Close()

		fmt.Println("Yeni bir kullanıcı bağlandı!")

		// Sürekli olarak mesajları dinle
		for {
			// Gelen mesajı oku
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Mesaj okunamadı:", err)
				break
			}

			// Gelen mesajı ekrana yaz
			fmt.Printf("Gelen mesaj: %s\n", message)

			// Aynı mesajı kullanıcıya geri gönder
			if err := conn.WriteMessage(messageType, message); err != nil {
				log.Println("Mesaj gönderilemedi:", err)
				break
			}
		}
	})

	// Sunucuyu başlat
	fmt.Println("WebSocket sunucusu 8080 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
