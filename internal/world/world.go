package world

import (
	"log"
	"math/rand"
	"sync"

	"mysimulation/internal/agent"
)

// Config — параметры мира
type Config struct {
	Width             float64 // Ширина мира по X
	Height            float64 // Высота мира по Z
	TickMs            int64   // Базовый тик в мс
	InitialPlants     int     // начальное количество растений
	InitialHerbivores int     // начальное количество травоядных
	InitialPredators  int     // начальное количество хищников
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Width:             20.0,
		Height:            20.0,
		TickMs:            50,
		InitialPlants:     25,
		InitialHerbivores: 12,
		InitialPredators:  4,
	}
}

// World — основной объект мира
type World struct {
	mu     sync.RWMutex
	agents map[int]*agent.Agent
	nextID int
	cfg    Config
}

// New создаёт новый мир с конфигурацией по умолчанию
func New() *World {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig создаёт новый мир с пользовательской конфигурацией
func NewWithConfig(cfg Config) *World {
	w := &World{
		agents: make(map[int]*agent.Agent),
		nextID: 1,
		cfg:    cfg,
	}
	w.Reset()
	return w
}

// Reset — очищает мир и создаёт начальную популяцию
func (w *World) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.resetLocked()
	log.Println("🌍 Мир перезапущен!")
}

// resetLocked — внутренний метод без захвата мьютекса
func (w *World) resetLocked() {
	w.agents = make(map[int]*agent.Agent)
	w.nextID = 1

	// Растения
	for i := 0; i < w.cfg.InitialPlants; i++ {
		w.addAgentLocked(agent.TypePlant, 80.0)
	}

	// Травоядные
	for i := 0; i < w.cfg.InitialHerbivores; i++ {
		w.addAgentLocked(agent.TypeHerbivore, 60.0)
	}

	// Хищники
	for i := 0; i < w.cfg.InitialPredators; i++ {
		w.addAgentLocked(agent.TypePredator, 80.0)
	}
}

// addAgentLocked — добавляет агента (без захвата мьютекса)
func (w *World) addAgentLocked(agentType string, energy float64) *agent.Agent {
	x := (rand.Float64() - 0.5) * w.cfg.Width
	z := (rand.Float64() - 0.5) * w.cfg.Height

	a := agent.New(w.nextID, agentType, x, 0.5, z, energy)
	w.agents[w.nextID] = a
	w.nextID++
	return a
}

// AddAgent — публичный метод добавления агента
func (w *World) AddAgent(agentType string, energy float64) *agent.Agent {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.addAgentLocked(agentType, energy)
}

// removeAgentLocked — удаляет агента (без захвата мьютекса)
func (w *World) removeAgentLocked(id int) {
	delete(w.agents, id)
}

// GetAgents — возвращает всех агентов (копию слайса)
func (w *World) GetAgents() []*agent.Agent {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]*agent.Agent, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, a)
	}
	return agentsList
}

// Update — один шаг симуляции
func (w *World) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Движение
	for _, a := range w.agents {
		w.moveAgentLocked(a)
	}

	// 2. Поедание
	for _, a := range w.agents {
		w.handleEatingLocked(a)
	}

	// 3. Трата энергии и смерть
	toRemove := []int{}
	for _, a := range w.agents {
		w.spendEnergyLocked(a)
		if a.Energy <= 0 {
			toRemove = append(toRemove, a.ID)
		}
	}
	for _, id := range toRemove {
		w.removeAgentLocked(id)
	}

	// 4. Размножение
	for _, a := range w.agents {
		w.tryReproduceLocked(a)
	}

	// 5. Восполнение растений
	w.growPlantsLocked()
}

// moveAgentLocked — движение агента
func (w *World) moveAgentLocked(a *agent.Agent) {
	if a.Type == agent.TypePlant {
		return
	}

	speed := 0.4
	if a.Type == agent.TypePredator {
		speed = 0.5
	}

	// Поиск цели
	var target *agent.Agent
	minDist := 100.0

	for _, other := range w.agents {
		if other == a {
			continue
		}

		isTarget := false
		switch a.Type {
		case agent.TypeHerbivore:
			isTarget = other.Type == agent.TypePlant
		case agent.TypePredator:
			isTarget = other.Type == agent.TypeHerbivore
		}

		if isTarget {
			dist := distance(a, other)
			if dist < minDist {
				minDist = dist
				target = other
			}
		}
	}

	if target != nil {
		dx := target.X - a.X
		dz := target.Z - a.Z
		length := dx*dx + dz*dz
		if length > 0 {
			dx /= length
			dz /= length
			a.X += dx * speed
			a.Z += dz * speed
		}
	} else {
		a.X += (rand.Float64() - 0.5) * speed
		a.Z += (rand.Float64() - 0.5) * speed
	}

	limit := w.cfg.Width/2 - 1
	if a.X > limit {
		a.X = limit
	}
	if a.X < -limit {
		a.X = -limit
	}
	if a.Z > limit {
		a.Z = limit
	}
	if a.Z < -limit {
		a.Z = -limit
	}
}

// handleEatingLocked — поедание
func (w *World) handleEatingLocked(a *agent.Agent) {
	eatRadius := 1.5

	switch a.Type {
	case agent.TypeHerbivore:
		var nearest *agent.Agent
		minDist := eatRadius

		for _, other := range w.agents {
			if other.Type == agent.TypePlant && other != a {
				dist := distance(a, other)
				if dist < minDist {
					minDist = dist
					nearest = other
				}
			}
		}

		if nearest != nil {
			w.removeAgentLocked(nearest.ID)
			a.Energy += 30
			if a.Energy > 100 {
				a.Energy = 100
			}
		}

	case agent.TypePredator:
		var nearest *agent.Agent
		minDist := eatRadius

		for _, other := range w.agents {
			if other.Type == agent.TypeHerbivore && other != a {
				dist := distance(a, other)
				if dist < minDist {
					minDist = dist
					nearest = other
				}
			}
		}

		if nearest != nil {
			w.removeAgentLocked(nearest.ID)
			a.Energy += 40
			if a.Energy > 100 {
				a.Energy = 100
			}
		}
	}
}

// spendEnergyLocked — трата энергии
func (w *World) spendEnergyLocked(a *agent.Agent) {
	cost := 0.5
	switch a.Type {
	case agent.TypeHerbivore:
		cost = 0.6
	case agent.TypePredator:
		cost = 0.8
	case agent.TypePlant:
		cost = 0.1
	}
	a.Energy -= cost
}

// tryReproduceLocked — размножение
func (w *World) tryReproduceLocked(a *agent.Agent) {
	requiredEnergy := 55.0
	if a.Energy < requiredEnergy {
		return
	}

	chance := 0.02
	switch a.Type {
	case agent.TypeHerbivore:
		chance = 0.04
	case agent.TypePredator:
		chance = 0.025
	case agent.TypePlant:
		chance = 0.05
	}

	if rand.Float64() < chance {
		a.Energy -= 25
		w.addAgentLocked(a.Type, 35.0)
	}
}

// growPlantsLocked — восстановление растений
func (w *World) growPlantsLocked() {
	plantCount := 0
	for _, a := range w.agents {
		if a.Type == agent.TypePlant {
			plantCount++
		}
	}

	minPlants := 15
	if plantCount < minPlants {
		newPlants := minPlants - plantCount
		for i := 0; i < newPlants; i++ {
			w.addAgentLocked(agent.TypePlant, 50.0)
		}
	}
}

// distance — квадрат расстояния между агентами
func distance(a, b *agent.Agent) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return dx*dx + dz*dz
}
