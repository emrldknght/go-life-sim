package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type Agent struct {
	ID    int     `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Type  string  `json:"type"`  // "herbivore", "predator", "plant"
	Color string  `json:"color"` // цвет для отображения
}

type World struct {
	agents map[int]*Agent
}

func NewWorld() *World {
	w := &World{
		agents: make(map[int]*Agent),
	}
	// Создаём 50 случайных агентов для теста
	for i := 0; i < 50; i++ {
		w.agents[i] = &Agent{
			ID:    i,
			X:     (rand.Float64() - 0.5) * 15, // -7.5 до 7.5
			Y:     0.5,
			Z:     (rand.Float64() - 0.5) * 15,
			Type:  "herbivore",
			Color: "#44ff44", // зелёный
		}
	}
	return w
}

func (w *World) GetAgentsJSON() []byte {
	agentsList := make([]*Agent, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, a)
	}
	data, _ := json.Marshal(agentsList)
	return data
}

func main() {
	world := NewWorld()

	// Запускаем симуляцию в отдельной горутине
	go simulationLoop(world)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, world)
	})
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Сервер на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func simulationLoop(world *World) {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 FPS
	for range ticker.C {
		// Пока просто двигаем агентов по кругу для теста
		for _, agent := range world.agents {
			agent.X += (rand.Float64() - 0.5) * 0.2
			agent.Z += (rand.Float64() - 0.5) * 0.2

			// Ограничиваем границы
			if agent.X > 10 {
				agent.X = 10
			}
			if agent.X < -10 {
				agent.X = -10
			}
			if agent.Z > 10 {
				agent.Z = 10
			}
			if agent.Z < -10 {
				agent.Z = -10
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request, world *World) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("WS accept error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	log.Println("Клиент подключился")
	ctx := context.Background()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		data := world.GetAgentsJSON()
		err := conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			log.Printf("Клиент отключился: %v", err)
			break
		}
	}
}
