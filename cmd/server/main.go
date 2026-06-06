package main

import (
	"log"
	"net/http"
	"time"

	"mysimulation/internal/websocket"
	"mysimulation/internal/world"
)

func main() {
	// Создаём мир
	w := world.New()

	// Создаём WebSocket хаб
	hub := websocket.NewHub(w)

	// Запускаем симуляцию в фоне
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		for range ticker.C {
			w.Update()
		}
	}()

	// Настраиваем маршруты
	http.HandleFunc("/ws", hub.Handler())
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
