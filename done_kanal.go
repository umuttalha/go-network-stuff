package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan bool) // `done` kanalı oluşturuluyor.

	go func() {
		fmt.Println("Goroutine başladı.")
		time.Sleep(1 * time.Second)
		done <- true // İş bittiğinde `done` kanalına bir değer gönderiliyor.
	}()

	<-done // `done` kanalından veri gelene kadar ana goroutine bekler.
	fmt.Println("Goroutine tamamlandı, program sona erdi.")
}
