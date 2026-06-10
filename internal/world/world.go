package world

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"mysimulation/internal/agent"
)

// Config — параметры мира
type Config struct {
	Width             float64 // Ширина мира по X
	Height            float64 // Высота мира по Z
	BaseTickMs        int64   // базовый тик в миллисекундах
	Speed             float64 // множитель скорости (0.25, 0.5, 1, 2, 4)
	InitialPlants     int     // начальное количество растений
	InitialHerbivores int     // начальное количество травоядных
	InitialPredators  int     // начальное количество хищников
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Width:             20.0,
		Height:            20.0,
		BaseTickMs:        50,
		Speed:             1.0,
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
	paused bool // добавляем флаг паузы
}

// AgentDTO — копия агента для безопасной сериализации
type AgentDTO struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Type   string  `json:"type"`
	Color  string  `json:"color"`
	Energy float64 `json:"energy"`
}

// GetAgentsDTO — возвращает безопасную копию всех агентов для JSON
func (w *World) GetAgentsDTO() []AgentDTO {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]AgentDTO, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, AgentDTO{
			ID:     a.ID,
			X:      a.X,
			Y:      a.Y,
			Z:      a.Z,
			Type:   a.Type,
			Color:  a.Color,
			Energy: a.Energy,
		})
	}
	return agentsList
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
		paused: false,
	}
	w.Reset()
	return w
}

// SetPaused — устанавливает состояние паузы
func (w *World) SetPaused(paused bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = paused
	if paused {
		log.Println("⏸️ Симуляция на паузе")
	} else {
		log.Println("▶️ Симуляция возобновлена")
	}
}

// IsPaused — возвращает состояние паузы
func (w *World) IsPaused() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.paused
}

// SetSpeed — устанавливает скорость симуляции
func (w *World) SetSpeed(speed float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Ограничиваем скорость разумными пределами
	if speed < 0.25 {
		speed = 0.25
	}
	if speed > 4.0 {
		speed = 4.0
	}

	w.cfg.Speed = speed
	log.Printf("⚡ Скорость симуляции изменена: x%.2f", speed)
}

// GetSpeed — возвращает текущую скорость
func (w *World) GetSpeed() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.cfg.Speed
}

// GetTickDuration — возвращает текущую длительность тика с учётом скорости
func (w *World) GetTickDuration() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tickMs := float64(w.cfg.BaseTickMs) / w.cfg.Speed
	if tickMs < 10 {
		tickMs = 10 // Минимум 10 мс (100 FPS симуляции)
	}
	return time.Duration(tickMs) * time.Millisecond
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
/* todo: - removed as unsafe
func (w *World) GetAgents() []*agent.Agent {

	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]*agent.Agent, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, a)
	}
	return agentsList
}
*/

// Update — один шаг симуляции
func (w *World) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.paused {
		return
	}

	// 1. Движение (скорость движения зависит от множителя)
	for _, a := range w.agents {
		w.moveAgentLocked(a)
	}

	// 2. Поедание (радиус не меняется, но частота обновлений уже выше)
	for _, a := range w.agents {
		w.handleEatingLocked(a)
	}

	// 3. Трата энергии и смерть (трата зависит от скорости)
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

	// 4. Размножение (шанс зависит от скорости)
	for _, a := range w.agents {
		w.tryReproduceLocked(a)
	}

	// 5. Восполнение растений (скорость роста зависит от скорости мира)
	w.growPlantsLocked()
}

// moveAgentLocked — движение агента с учётом скорости мира
func (w *World) moveAgentLocked(a *agent.Agent) {
	if a.Type == agent.TypePlant {
		return
	}

	// Базовая скорость агента
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
		// ИСПРАВЛЕНО: используем math.Sqrt для получения реального расстояния
		length := math.Sqrt(dx*dx + dz*dz)
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

// spendEnergyLocked — трата энергии с учётом скорости мира
func (w *World) spendEnergyLocked(a *agent.Agent) {
	// Базовая стоимость
	baseCost := 0.5
	switch a.Type {
	case agent.TypeHerbivore:
		baseCost = 0.6
	case agent.TypePredator:
		baseCost = 0.8
	case agent.TypePlant:
		baseCost = 0.1
	}
	a.Energy -= baseCost
}

// tryReproduceLocked — размножение с учётом скорости мира
func (w *World) tryReproduceLocked(a *agent.Agent) {
	requiredEnergy := 55.0
	if a.Energy < requiredEnergy {
		return
	}

	// Базовый шанс
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

// growPlantsLocked — восстановление растений с учётом скорости
func (w *World) growPlantsLocked() {
	plantCount := 0
	for _, a := range w.agents {
		if a.Type == agent.TypePlant {
			plantCount++
		}
	}

	// Целевое количество растений
	targetPlants := 15

	if plantCount < targetPlants {
		// Количество новых растений зависит от скорости
		newPlants := targetPlants - plantCount
		if newPlants < 1 {
			newPlants = 1
		}
		if newPlants > 10 {
			newPlants = 10
		}
		for i := 0; i < newPlants; i++ {
			w.addAgentLocked(agent.TypePlant, 50.0)
		}
	}
}

// distance — реальное расстояние между агентами
func distance(a, b *agent.Agent) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dz*dz)
}
