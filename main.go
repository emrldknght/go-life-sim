package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const (
	TypePlant     = "plant"
	TypeHerbivore = "herbivore"
	TypePredator  = "predator"
)

var colors = map[string]string{
	TypePlant:     "#44ff44",
	TypeHerbivore: "#44aaff",
	TypePredator:  "#ff4444",
}

type Agent struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Type   string  `json:"type"`
	Color  string  `json:"color"`
	Energy float64 `json:"energy"`
}

type World struct {
	mu     sync.RWMutex
	agents map[int]*Agent
	nextID int
	width  float64
	height float64
}

func NewWorld() *World {
	w := &World{
		agents: make(map[int]*Agent),
		nextID: 1,
		width:  20.0,
		height: 20.0,
	}
	w.Reset()
	return w
}

// Reset — очищает мир и создаёт начальную популяцию
func (w *World) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Очищаем всё
	w.agents = make(map[int]*Agent)
	w.nextID = 1

	// Создаём начальную популяцию
	// Растения (25 штук)
	for i := 0; i < 25; i++ {
		w.addAgentLocked(TypePlant, 80.0)
	}

	// Травоядные (12 штук)
	for i := 0; i < 12; i++ {
		w.addAgentLocked(TypeHerbivore, 60.0)
	}

	// Хищники (4 штуки)
	for i := 0; i < 4; i++ {
		w.addAgentLocked(TypePredator, 80.0)
	}

	log.Println("🌍 Мир перезапущен!")
}

// addAgentLocked — без захвата мьютекса (для внутреннего использования)
func (w *World) addAgentLocked(agentType string, energy float64) *Agent {
	agent := &Agent{
		ID:     w.nextID,
		X:      (rand.Float64() - 0.5) * w.width,
		Y:      0.5,
		Z:      (rand.Float64() - 0.5) * w.height,
		Type:   agentType,
		Color:  colors[agentType],
		Energy: energy,
	}
	w.nextID++
	w.agents[agent.ID] = agent
	return agent
}

// AddAgent — публичный метод с мьютексом
func (w *World) AddAgent(agentType string, energy float64) *Agent {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.addAgentLocked(agentType, energy)
}

func (w *World) removeAgentLocked(id int) {
	delete(w.agents, id)
}

func (w *World) GetAgentsJSON() []byte {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]*Agent, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, a)
	}
	data, _ := json.Marshal(agentsList)
	return data
}

func (w *World) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Движение
	for _, agent := range w.agents {
		w.moveAgentLocked(agent)
	}

	// 2. Поедание
	for _, agent := range w.agents {
		w.handleEatingLocked(agent)
	}

	// 3. Трата энергии и смерть
	toRemove := []int{}
	for _, agent := range w.agents {
		w.spendEnergyLocked(agent)
		if agent.Energy <= 0 {
			toRemove = append(toRemove, agent.ID)
		}
	}
	for _, id := range toRemove {
		w.removeAgentLocked(id)
	}

	// 4. Размножение
	for _, agent := range w.agents {
		w.tryReproduceLocked(agent)
	}

	// 5. Восполнение растений
	w.growPlantsLocked()
}

func (w *World) moveAgentLocked(agent *Agent) {
	if agent.Type == TypePlant {
		return
	}

	speed := 0.4
	if agent.Type == TypePredator {
		speed = 0.5
	}

	// Поиск цели
	var target *Agent
	minDist := 100.0

	for _, other := range w.agents {
		if other == agent {
			continue
		}

		var isTarget bool
		switch agent.Type {
		case TypeHerbivore:
			isTarget = other.Type == TypePlant
		case TypePredator:
			isTarget = other.Type == TypeHerbivore
		}

		if isTarget {
			dist := distance(agent, other)
			if dist < minDist {
				minDist = dist
				target = other
			}
		}
	}

	if target != nil {
		dx := target.X - agent.X
		dz := target.Z - agent.Z
		length := dx*dx + dz*dz
		if length > 0 {
			dx /= length
			dz /= length
			agent.X += dx * speed
			agent.Z += dz * speed
		}
	} else {
		agent.X += (rand.Float64() - 0.5) * speed
		agent.Z += (rand.Float64() - 0.5) * speed
	}

	limit := 9.0
	if agent.X > limit {
		agent.X = limit
	}
	if agent.X < -limit {
		agent.X = -limit
	}
	if agent.Z > limit {
		agent.Z = limit
	}
	if agent.Z < -limit {
		agent.Z = -limit
	}
}

func (w *World) handleEatingLocked(agent *Agent) {
	switch agent.Type {
	case TypeHerbivore:
		var nearestPlant *Agent
		minDist := 1.5

		for _, other := range w.agents {
			if other.Type == TypePlant && other != agent {
				dist := distance(agent, other)
				if dist < minDist {
					minDist = dist
					nearestPlant = other
				}
			}
		}

		if nearestPlant != nil {
			w.removeAgentLocked(nearestPlant.ID)
			agent.Energy += 30
			if agent.Energy > 100 {
				agent.Energy = 100
			}
		}

	case TypePredator:
		var nearestPrey *Agent
		minDist := 1.5

		for _, other := range w.agents {
			if other.Type == TypeHerbivore && other != agent {
				dist := distance(agent, other)
				if dist < minDist {
					minDist = dist
					nearestPrey = other
				}
			}
		}

		if nearestPrey != nil {
			w.removeAgentLocked(nearestPrey.ID)
			agent.Energy += 40
			if agent.Energy > 100 {
				agent.Energy = 100
			}
		}
	}
}

func (w *World) spendEnergyLocked(agent *Agent) {
	cost := 0.5
	switch agent.Type {
	case TypeHerbivore:
		cost = 0.6
	case TypePredator:
		cost = 0.8
	case TypePlant:
		cost = 0.1
	}
	agent.Energy -= cost
}

func (w *World) tryReproduceLocked(agent *Agent) {
	requiredEnergy := 55.0
	if agent.Energy < requiredEnergy {
		return
	}

	chance := 0.02
	switch agent.Type {
	case TypeHerbivore:
		chance = 0.04
	case TypePredator:
		chance = 0.025
	case TypePlant:
		chance = 0.05
	}

	if rand.Float64() < chance {
		agent.Energy -= 25
		childEnergy := 35.0
		w.addAgentLocked(agent.Type, childEnergy)
	}
}

func (w *World) growPlantsLocked() {
	plantCount := 0
	for _, agent := range w.agents {
		if agent.Type == TypePlant {
			plantCount++
		}
	}

	if plantCount < 15 {
		newPlants := 15 - plantCount
		for i := 0; i < newPlants; i++ {
			w.addAgentLocked(TypePlant, 50.0)
		}
	}
}

func distance(a, b *Agent) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return dx*dx + dz*dz
}

func main() {
	world := NewWorld()

	// Симуляция
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		for range ticker.C {
			world.Update()
		}
	}()

	// WebSocket с поддержкой команд от клиента
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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

		// Канал для отправки данных клиенту
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		// Канал для команд от клиента
		go func() {
			for {
				_, msg, err := conn.Read(ctx)
				if err != nil {
					log.Printf("Клиент отключился: %v", err)
					break
				}

				// Обработка команд
				var command map[string]string
				if err := json.Unmarshal(msg, &command); err == nil {
					if command["action"] == "reset" {
						world.Reset()
						log.Println("🔄 Получена команда перезапуска")
					}
				}
			}
		}()

		// Отправка данных клиенту
		for range ticker.C {
			data := world.GetAgentsJSON()
			err := conn.Write(ctx, websocket.MessageText, data)
			if err != nil {
				break
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("🚀 Сервер на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
